// hsterm terminal thing with history, complete (not done yet)
// needs code cleanup, maybe stop the clear string instead of redraw

package termu

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

	"github.com/gohxs/prettylog"
	"github.com/gohxs/termu/ansi"
	"github.com/gohxs/termu/term"
	"github.com/gohxs/termu/term/termutils"
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
	writer, err := term.NewStdoutWriter(os.Stdout)
	if err != nil {
		writer = os.Stdout // pure stdout?

	}

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
		outWriter: os.Stdout, // Why?

		// Accept readwriter like golang.org/x/crypt/ssh/terminal
		Reader: term.NewStdinReader(os.Stdin), // Reader
		Writer: writer,                        // TODO: for now
		//m:      sync.Mutex{},
	}

	return ret
}

// ReadLine will wait for further input until results is satisfie
// (i.e: a line is readed)
func (t *Term) ReadLine() (string, error) {
	t.width, _ = t.GetSize() // Update width actually

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

	reader := t.Reader          //erm.NewStdinReader(ifile) // Wrapper for windows/unix
	t.out.WriteString("\033[s") // Save cursor now?
	t.out.WriteString(t.promptString())
	t.Flush()

	//
	rinput := ansi.NewReader(reader)

	for {
		t.width, _ = t.GetSize() // Update width actually
		val, err := rinput.ReadEscape()
		if err != nil {
			return "", err
		}
		// Do some keyMapping, (i.e: MoveNext: "\t")
		// Plugable Modes like the readline

		// Select handler here
		//
		// Non clear
		switch val.Value {
		case "\x15": // CtrlU
			t.Write([]byte("\033[H"))
		case "\x0C": // CtrlL
			ab := ansi.NewHelperDirect(&t.out)
			ab.WriteString("\033[2J\033[H") // Clear and save?
			ab.SaveCursor()
			t.Flush()

		case "\x04": //EOT // CtrlD
			t.out.WriteString(t.promptString()) // Get String
			continue
		case "\x03": // CtrlC
			return "", ErrInterrupt // Interrupt return
		case "\r", "\n": // ENTER COMPLETE enter // Process input
			//t.inbuf.CursorToEnd() // Index to end of prompt what if?
			t.addLine = 0
			t.out.WriteString("\n")                   // Line feed directly
			ansi.NewHelperDirect(&t.out).SaveCursor() // ansi for just save cursor?
			t.Flush()                                 // Send to terminal
			if t.inbuf.Len() == 0 {
				t.out.WriteString(t.promptString()) // Reprint prompt
				t.Flush()                           // Send to terminal
				continue
			}
			// Append history
			t.History.Append(t.inbuf.String())

			line := t.inbuf.String()
			t.inbuf.Clear()
			return line, nil
		}

		switch val.Value {
		case "\033f": // Word back
			t.inbuf.CursorWordForward()
		case "\033b": // Word back
			t.inbuf.CursorWordBack()
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
		}
		// Lock here
		t.out.WriteString(t.promptString())
		t.Flush()
	}
	log.Println("Readline exited")
	return "", nil
}

//Write can be called by any other thread and a Flush triggered
func (t *Term) Write(b []byte) (n int, err error) {
	if len(b) == 0 {
		return 0, nil
	}
	t.addLine = 0
	if b[len(b)-1] != '\n' { // Prompt will show below the last cursor in next print
		t.addLine = 1
	}
	if !t.reading { //Pass through
		n, err = t.out.Write(b)
		t.Flush() // Cursor thing
		return
	}

	ab := ansi.NewHelperBuffered(&t.out)

	ab.RestoreCursor() // Restoring cursor does not work in some terminals
	// XXX: Bug, this prevent us to manipulate the lines while writing
	// Ideally we would clear the prompt only
	ab.WriteString("\033[K") // Clear after restore? before? Wrong
	ab.Write(b)
	if len(b) == 1 {
		ab.WriteString(" \b") // Linux things to force line change
	}
	ab.SaveCursor()

	ab.WriteString(t.promptString())
	ab.Flush()
	t.Flush()

	return
}

//GetSize get terminal size returns 0,0 on error
func (t *Term) GetSize() (w int, h int) {
	f, ok := t.outWriter.(*os.File)
	if !ok {
		return
	}

	iw, ih, err := termutils.GetSize(int(f.Fd()))
	if err != nil {
		return
	}

	return iw, ih
}

//SetPrompt sets the terminal prompt
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

//XXX: Private
// Reset Prompt Cursor
func (t *Term) CleanPromptString() string {
	return ""
}

//TODO: This should be private
//PromptString Returns the prompt string proper escaped
func (t *Term) promptString() string {
	ab := ansi.NewHelperBuffered(nil)
	ab.RestoreCursor() // Or prompt clear
	//For redraw
	count := t.PromptLineCount() - 1 + t.addLine
	if count > 0 { // Move back input and thing buffer?
		// Scroll hack
		// with restore cursor, if it is at last line,
		// it will scroll up any remaining text repeating the top line of the prompt
		// We line feed the number of lines prompt will have
		// but since the cursor will go to 0 position, we have to:
		// restore again,
		// As the cursor is at bottom line, so we moveDown (will not move if in bottom)
		// and move up back to desired position, saving the cursor for the next Write
		// if the cursor is not on the last line it will perform movements and
		// will end up same position
		ab.WriteString(strings.Repeat("\n", count)) // Form feed if necessary // implemente on windows somehow
		ab.RestoreCursor()                          // Move back
		ab.MoveDown(count)                          //
		ab.MoveUp(count)
		ab.SaveCursor() // Save cursor again
	} /**/

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

	// Here we are printing display
	ab.WriteString("\r" + t.prompt)
	ab.WriteString(dispBuf)

	// If unix only
	if term.Variant == term.VariantUnix {
		ab.WriteString(" \b")
	}

	//XXX: if we add a special printer that prints bellow
	//screen it will becleared, so we need to only clear the LINE2
	//ab.WriteString("\033[K") // Clean Rest of the line
	ab.WriteString("\033[J") // Maybe rest of the line
	// Position cursor // Reposition cursor here??
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

	ab.MoveUp(lineCount) // Origin
	ab.WriteString("\r") // go back anyway
	ab.MoveDown(desiredLine)
	ab.MoveRight(desiredCol)

	/*dispLen := t.inbuf.Len()
	curLeft := (dispLen - t.inbuf.Cursor())
	if curLeft != 0 && dispLen > 0 {
		ab.WriteString(fmt.Sprintf("\033[%dD", curLeft)) // ? huh
	}*/

	return ab.String()

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
