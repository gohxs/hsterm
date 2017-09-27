// hsterm terminal thing with history, complete (not done yet)
// needs code cleanup, maybe stop the clear string instead of redraw

package hsterm

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/gohxs/hsterm/ansi"
	"github.com/gohxs/hsterm/term"
	"github.com/gohxs/hsterm/term/termutils"
	"github.com/gohxs/prettylog"
)

var (
	// ErrInterrupt Interrupt called
	ErrInterrupt = errors.New("Interrupt")
)

// TermFunc callback for receiving a command
//type termFunc func(string)

// Term main var
type Term struct {
	Display      func(string) string
	AutoComplete func(line string, pos int, key rune) (newLine string, newPos int, ok bool)
	// Internals
	History History // Expose history
	inbuf   *InputBuffer
	prompt  string
	width   int

	Log    *log.Logger
	out    bytes.Buffer // Buffered output
	tstate *termutils.State

	reading bool
	addLine int

	inReader  io.Reader
	outWriter io.Writer

	Reader io.Reader
	Writer io.Writer
	// Stderr in future

	//inFile *os.File // io.Reader
	//m sync.Mutex
}

//New instantiates a new Terminal handler
func New() *Term {

	//inFile := term.Stdin
	//outFile := term.Stdout
	// For windows only
	//inFD := syscall.STD_INPUT_HANDLE
	//outFD := syscall.STD_OUTPUT_HANDLE

	ret := &Term{
		Display:      nil,
		AutoComplete: nil,
		// Internals
		inbuf:   NewInputBuffer(),
		prompt:  "",
		width:   0,
		History: &history{},
		Log:     prettylog.Dummy(),

		// TODO: Way to change this
		inReader:  os.Stdin,
		outWriter: os.Stdout,

		// Accept readwriter like golang.org/x/crypt/ssh/terminal
		Reader: term.NewStdinReader(os.Stdin),   // Reader
		Writer: term.NewStdoutWriter(os.Stdout), // TODO: for now
		//m:      sync.Mutex{},
	}

	return ret
}

// ReadLine will wait for further input until results is satisfie
// (i.e: a line is readed)
func (t *Term) ReadLine() (string, error) {
	t.width = t.GetWidth() // Update width actually

	ifile, ok := t.inReader.(*os.File)
	if !ok || (ok && !termutils.IsTerminal(int(ifile.Fd()))) {
		reader := bufio.NewReader(t.Reader)
		buf, _, err := reader.ReadLine()
		return string(buf), err
	}

	// It is a file
	state, err := termutils.MakeRaw(int(ifile.Fd()))
	if err != nil {
		panic(err)
	}
	defer termutils.Restore(int(ifile.Fd()), state)

	t.reading = true
	defer func() { t.reading = false }()

	// Get A STDIN Reader here windows is different
	// Prepare a reader and writer here

	reader := t.Reader //erm.NewStdinReader(ifile) // Wrapper for windows/unix
	//	t.m.Lock()
	{ // Terminal must support Ansi codes, i.e: writer Wrapper
		t.out.WriteString("\033[s") // Save cursor now?
		t.out.WriteString(t.PromptString())
		t.Flush()
	}
	//	t.m.Unlock()

	//
	rinput := ansi.NewReader(reader)

	for {
		t.width = t.GetWidth() // Update width actually
		val, err := rinput.ReadEscape()
		if err != nil {
			return "", err
		}
		// Do some keyMapping, (i.e: MoveNext: "\t")

		// Select handler here
		//
		// Non clear
		switch val.Value {
		case "\x15": // CtrlU
			t.Write([]byte("\033[H"))
		case "\x0C": // CtrlL
			ab := ansi.NewHelper(&t.out)
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
				ansi.NewHelper(&t.out).SaveCursor() // ansi for just save cursor?
				t.Flush()                           // Send to terminal
			}
			//			t.m.Unlock()
			if t.inbuf.Len() == 0 {
				t.out.WriteString(t.PromptString()) // Reprint prompt
				t.Flush()                           // Send to terminal
				continue
			}
			// Append history
			t.History.Append(t.inbuf.String())

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
		case "\033[A", "\x10": // Up or CtrlP
			t.inbuf.Set(t.History.Prev(t.inbuf.String()))
		case "\033[B", "\x0E": // Down or CtrlN
			t.inbuf.Set(t.History.Next(t.inbuf.String()))
		case "\033[D", "\x02": // left or CtrlB
			t.inbuf.CursorLeft()
		case "\033[C", "\x06": // Right or ctrlF
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
	ab := ansi.NewHelper(buf)

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

func (t *Term) GetWidth() int {
	f, ok := t.outWriter.(*os.File)
	if !ok {
		return -1
	}
	w, _, err := termutils.GetSize(int(f.Fd()))
	if err != nil {
		return -1
	}

	return w
}
func (t *Term) GetHeight() int {
	f, ok := t.outWriter.(*os.File)
	if !ok {
		return -1
	}
	_, h, err := termutils.GetSize(int(f.Fd()))
	if err != nil {
		return -1
	}
	return h
}

// SetPrompt sets the terminal prompt
func (t *Term) SetPrompt(p string) {
	t.prompt = p
}

//PromptLineCount helper to count wrapping lines
func (t *Term) PromptLineCount() int {
	//Count \n chars too
	width := 1
	if t.width > 0 {
		width = t.width
	}

	return ((t.inbuf.Len() + len(t.prompt)) / width) + 1 // Always one line
}

// PromptString Returns the prompt string proper escaped
func (t *Term) PromptString() string {
	buf := bytes.NewBuffer(nil)
	ab := ansi.NewHelper(buf)

	ab.RestoreCursor()
	//For redraw
	//t.RestoreCursor() // Last output position
	count := t.PromptLineCount() - 1 + t.addLine

	if count > 0 { // Move back input and thing buffer?
		ab.WriteString(strings.Repeat("\n", count)) // Form feed if necessary
		ab.RestoreCursor()
		ab.MoveDown(count) // Trick to positioning cursor?
		ab.MoveUp(count)
		//ab.MoveDown(
		//ab.WriteString(fmt.Sprintf("\033[%dA", count)) // Go down, go up, and left

		//ab.WriteString(fmt.Sprintf("\033[%dA", count)) // Go down, go up, and left
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

	// If unix only
	if term.Variant == term.VariantUnix {
		ab.WriteString(" \b")
	}
	ab.WriteString("\033[J") // Clean Rest of the screen

	//ab.WriteString("\033[J") // Maybe rest of the line
	//fmt.Sprintf("%s%s \b\033[J", t.prompt, dispBuf)) // Erase after prompt?

	// Position cursor // Reposition cursor here??
	//

	// No unicode for now
	width := min(t.width, 1)

	inbufLen := t.inbuf.Len() // Should count rune width
	fullWidth := len(t.prompt) + inbufLen

	cursorWidth := fullWidth - (inbufLen - t.inbuf.Cursor())
	lineCount := count

	desiredLine := (cursorWidth / width) // get Line position starting from prompt
	desiredCol := cursorWidth % width    // get column position starting from prompt

	// Go back instead of up
	lineCount -= t.addLine

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
	t.Log.Printf("Raw: %#v", t.out.String())

	writer := t.Writer

	writer.Write([]byte("\033[?25l")) // Hide cursor?
	writer.Write(t.out.Bytes())       // output
	writer.Write([]byte("\033[?25h")) // Show cursor?
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
		writer.Write([]byte(val.Raw))
		switch val.Value {
		case "\033[u":
			t.Log.Println("Restoring cursor")
			<-time.After(200 * time.Millisecond)
		case "\033[s":
			t.Log.Println("Saving cursor")
			<-time.After(200 * time.Millisecond)
		default:
			t.Log.Printf("%#v", val.Raw)
		}
		<-time.After(146 * time.Millisecond)

	}
	//log.Println("Reset buffer")
	t.out.Reset()

	// Flush into screen
}
