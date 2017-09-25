// hsterm terminal thing with history, complete (not done yet)
// needs code cleanup, maybe stop the clear string instead of redraw

package hsterm

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	glog "log"
	"os"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/gohxs/hsterm/ansi"
	"github.com/gohxs/hsterm/term"
	"github.com/gohxs/prettylog"
)

var (
	// ErrInterrupt Interrupt called
	ErrInterrupt = errors.New("Interrupt")
	// Local logger
	log         = glog.New(os.Stderr, "", 0)
	DebugOutput = ""
)

// TermFunc callback for receiving a command
//type termFunc func(string)

// Term main var
type Term struct {
	Display      func(string) string
	AutoComplete func(line string, pos int, key rune) (newLine string, newPos int, ok bool)
	// Internals
	history History
	inbuf   *InputBuffer
	prompt  string
	width   int

	out    bytes.Buffer // Buffered output
	tstate *term.State

	reading bool
	addLine int

	Reader io.Reader
	Writer io.Writer
	// Stderr in future

	//inFile *os.File // io.Reader
	//m sync.Mutex
}

//New instantiates a new Terminal handler
func New() *Term {

	inFile := term.Stdin
	outFile := term.Stdout

	// For windows only
	//inFD := syscall.STD_INPUT_HANDLE
	//outFD := syscall.STD_OUTPUT_HANDLE

	ret := &Term{
		Display:      nil,
		AutoComplete: nil,
		// Internals
		inbuf:   NewInputBuffer(),
		prompt:  "",
		width:   term.GetScreenWidth(),
		history: History{},

		Reader: inFile,  // Reader
		Writer: outFile, // TODO: for now
		//m:      sync.Mutex{},
	}

	{ // Advancing log/tmux helper
		log.Println("Opening the thing")
		f, err := os.OpenFile(DebugOutput, os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.FileMode(0644))
		if err != nil {
			panic(err)
		}
		log = prettylog.New("", f) // Debug logger
		ret.history.Append("llllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllllll")
	}
	return ret
}

// can be calledi n other thread and a Flush triggered
func (t *Term) Write(b []byte) (n int, err error) {
	if len(b) == 0 {
		return 0, nil
	}
	/*t.m.Lock()
	defer t.m.Unlock()*/
	// Write to temporary buffer
	t.addLine = 0
	if b[len(b)-1] != '\n' { // Write enter if inexistent
		t.addLine = 1
	}

	if !t.reading { // Pure write
		n, err = t.out.Write(b)
		t.Flush()
		return
	}

	buf := bytes.NewBuffer(nil)
	ab := ansi.NewWriter(buf)

	ab.RestoreCursor()
	ab.WriteString("\033[K") // Clear after restore
	ab.Write(b)
	if len(b) == 1 {
		ab.WriteString(" \b")
	}

	ab.SaveCursor()

	ab.WriteString(t.PromptString())
	// Finish
	t.out.Write(buf.Bytes())
	t.Flush()

	return
}

// SetPrompt sets the terminal prompt
func (t *Term) SetPrompt(p string) {
	t.prompt = p
}

// ReadLine will wait for further input until results is satisfie
// (i.e: a line is readed)
func (t *Term) ReadLine() (string, error) {

	ifile, ok := t.Reader.(*os.File)
	if !ok || (ok && !term.IsTerminal(int(ifile.Fd()))) {
		reader := bufio.NewReader(t.Reader)
		buf, _, err := reader.ReadLine()
		return string(buf), err
	}
	// It is a file
	state, err := term.MakeRaw(int(ifile.Fd()))
	if err != nil {
		panic(err)
	}

	defer term.Restore(int(ifile.Fd()), state)

	//	t.m.Lock()
	{
		t.out.WriteString("\033[s") // Save cursor now?
		t.out.WriteString(t.PromptString())
		t.Flush()
	}
	//	t.m.Unlock()

	t.reading = true
	defer func() { t.reading = false }()

	rinput := ansi.NewReader(t.Reader)
	for {
		t.width = term.GetScreenWidth() // Here?
		val, err := rinput.ReadEscape()
		if err != nil {
			return "", err
		}
		// Do some keyMapping, (i.e: MoveNext: "\t")

		// Non clear
		switch val.Value {
		case "\x15": // CtrlU
			t.Write([]byte("\033[H"))
		case "\x0C": // CtrlL
			ab := ansi.NewWriter(&t.out)
			ab.WriteString("\033[2J\033[H") // Clear and save?
			ab.SaveCursor()
			t.Flush()
			//continue
		case "\x04": //EOT // CtrlD
			//t.ClearPrompt()
			t.out.WriteString(t.PromptString()) // Get String
			continue
		case "\x03": // CtrlC
			return "", ErrInterrupt // Interrupt return
		case "\r", "\n": // ENTER COMPLETE enter // Process input
			// Process thing somewhere else
			t.inbuf.CursorToEnd() // Index to end of prompt what if?
			//			t.m.Lock()
			{
				t.addLine = 0
				t.out.WriteString("\n")             // Line feed directly
				ansi.NewWriter(&t.out).SaveCursor() // ansi for just save cursor?
				t.Flush()                           // Send to terminal
			}
			//			t.m.Unlock()
			if t.inbuf.Len() == 0 {
				t.out.WriteString(t.PromptString()) // Reprint prompt
				t.Flush()                           // Send to terminal
				continue
			}
			// Append history
			t.history.Append(t.inbuf.String())

			line := t.inbuf.String()
			t.inbuf.Clear()
			//log.Println("Return line")
			return line, nil
		}

		//t.RestoreCursor()
		switch val.Value {
		case "\033[3~": // Escape key delete Weird keys
			t.inbuf.Delete()
		case "\b", "\x7f":
			t.inbuf.Backspace()
		case "\033[A":
			t.inbuf.Set(t.history.Prev())
		case "\033[B":
			t.inbuf.Set(t.history.Next())
		case "\033[D":
			t.inbuf.CursorLeft()
		case "\033[C":
			t.inbuf.CursorRight()
		default: // Function that actually adds the text
			if val.Type != ansi.TypeRune {
				continue
			}

			ch, _ := utf8.DecodeRuneInString(val.Value)
			// Go through auto complete
			complete := false
			if t.AutoComplete != nil {
				newLine, newPos, ok := t.AutoComplete(t.inbuf.String(), t.inbuf.Cursor(), ch)
				if ok {
					t.inbuf.Set(newLine) // Reset print here?
					t.inbuf.SetCursor(newPos)
					complete = true
					//break
				}
			}
			if !complete && unicode.IsPrint(ch) { // Do not add the tab
				t.inbuf.WriteString(val.Value)
			}
			//Flush what?
			//t.Flush()
		}
		// Lock here
		//t.ResetPromptCursor()
		//t.m.Lock()
		t.out.WriteString(t.PromptString())
		t.Flush()

		// Internal type should never be exposedt.Flush()
		//t.m.Unlock()

		//t.m.Unlock()
	}
	log.Println("Readline exited")
	return "", nil
}

//PromptLineCount helper to count wrapping lines
func (t *Term) PromptLineCount() int {
	//Count \n chars too
	return ((t.inbuf.Len() + len(t.prompt)) / t.width) + 1 // Always one line
}

// PromptString Returns the prompt string proper escaped
func (t *Term) PromptString() string {
	buf := bytes.NewBuffer(nil)
	ab := ansi.NewWriter(buf)

	ab.RestoreCursor()
	//For redraw
	//t.RestoreCursor() // Last output position
	count := t.PromptLineCount() - 1 + t.addLine

	if count > 0 { // Move back input and thing buffer?
		ab.WriteString(fmt.Sprintf("%s\033[%dA", strings.Repeat("\f", count), count)) // Go down, go up, and left
	}
	ab.SaveCursor() // Save cursor again

	// Transform output
	var dispBuf string
	// Idea cache this, every time we print the thing, we transform
	// if transformation is different than cache we redraw, else we put the difference
	if t.Display != nil {
		dispBuf = t.Display(t.inbuf.String())
	} else {
		dispBuf = t.inbuf.String()
	}

	ab.WriteString(strings.Repeat("\n", t.addLine))

	ab.WriteString(strings.Repeat("\033[K", t.PromptLineCount())) // N lines

	ab.WriteString(t.prompt)
	ab.WriteString(dispBuf)
	ab.WriteString("\033[J") // Clean Rest of the screen

	//ab.WriteString("\033[J") // Maybe rest of the line
	//fmt.Sprintf("%s%s \b\033[J", t.prompt, dispBuf)) // Erase after prompt?

	// Position cursor // Reposition cursor here??
	//

	// No unicode for now
	inbufLen := t.inbuf.Len() // Should count rune width
	fullWidth := len(t.prompt) + inbufLen

	cursorWidth := fullWidth - (inbufLen - t.inbuf.Cursor())
	lineCount := count

	desiredLine := (cursorWidth / t.width) // get Line position starting from prompt
	desiredCol := cursorWidth % t.width    // get column position starting from prompt

	// Go back instead of up
	lineCount -= t.addLine
	log.Println("Cursor width:", cursorWidth)

	ab.MoveUp(lineCount)
	ab.WriteString("\r") // go back anyway
	ab.MoveDown(desiredLine)
	ab.MoveRight(desiredCol)

	/*dispLen := t.inbuf.Len()
	curLeft := (dispLen - t.inbuf.Cursor())
	if curLeft != 0 && dispLen > 0 {
		ab.WriteString(fmt.Sprintf("\033[%dD", curLeft)) // ? huh
	}*/

	return buf.String()

}

var flushMutex sync.Mutex

// Flush disables cursor renders and enable
func (t *Term) Flush() {
	flushMutex.Lock()
	defer func() {
		flushMutex.Unlock()
		//log.Println("Flushed")
	}()
	log.Printf("Raw: %#v", t.out.String())

	t.Writer.Write([]byte("\033[?25l")) // Hide cursor?
	t.Writer.Write(t.out.Bytes())       // output
	t.Writer.Write([]byte("\033[?25h")) // Show cursor?
	t.out.Reset()
	return /**/

	// DEBUG AREA:
	// lock while flushing
	//
	// Debugger
	rd := ansi.NewReader(bytes.NewReader(t.out.Bytes()))

	for {
		val, err := rd.ReadEscape()
		if err != nil {
			break
		}
		t.Writer.Write([]byte(val.Raw))
		switch val.Value {
		case "\033[u":
			log.Println("Restoring cursor")
			<-time.After(200 * time.Millisecond)
		case "\033[s":
			log.Println("Saving cursor")
			<-time.After(200 * time.Millisecond)
		default:
			//log.Printf("Normal escape :%#v", val.Raw)
		}
		<-time.After(46 * time.Millisecond)

	}
	//log.Println("Reset buffer")
	t.out.Reset()

	// Flush into screen
}
