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
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/gohxs/prettylog"
	"github.com/gohxs/termu/ansi"
	"github.com/gohxs/termu/termutils"
)

// Errors
var (
	ErrInterrupt = errors.New("Interrupt")
	ErrEOF       = errors.New("EOF")
)

// Virtual keys
const (
	AKCtrlA       rune = 1
	AKInterrupt        = 0x3
	AKeot              = 0x4
	AKBackspace        = 0x8
	AKEnter            = 0xD
	AKClearScreen      = 0xC

	AKDelete = 0xED00 + iota
	AKWordForward
	AKWordBack
	AKHistPrev
	AKHistNext
	AKCursLeft
	AKCursRight
	AKShiftTab
)

// TermFunc callback for receiving a command
//type termFunc func(string)

// Term main var
type Term struct {
	Display      func(string) string
	AutoComplete func(line string, pos int, key rune) (newLine string, newPos int, ok bool)
	// Internals
	History History // Expose history
	prompt  *Prompt
	width   int

	Log    *log.Logger
	out    bytes.Buffer // Buffered output
	tstate *termutils.State

	reading bool

	Reader io.Reader
	Writer io.Writer
}

//New instantiates a new Terminal handler
func New() *Term {
	// Setup the Read and Writer
	writer, err := termutils.NewStdoutWriter(os.Stderr)
	if err != nil {
		writer = os.Stdout
	}

	reader := io.Reader(termutils.NewStdinReader(os.Stdin))
	ifile, ok := reader.(*os.File)
	// if is not a file or is a file but not a terminal
	// We set our reader as a bufio.Reader
	if !ok || (ok && !termutils.IsTerminal(int(ifile.Fd()))) {
		reader = bufio.NewReader(reader)
	}

	ret := &Term{
		Display:      nil,
		AutoComplete: nil,
		// Internals
		width:   1,
		History: &history{}, // Wrong why?
		Log:     prettylog.Dummy(),

		// Accept readwriter like golang.org/x/crypt/ssh/terminal
		Reader: reader,
		Writer: writer,
	}
	ret.prompt = newPrompt(ret)

	return ret
}

// ReadLine will wait for further input until results is satisfie
// (i.e: a line is readed)
func (t *Term) ReadLine() (string, error) {

	// If it is the bufio.Reader we pass through
	if reader, ok := t.Reader.(*bufio.Reader); ok {
		buf, _, err := reader.ReadLine()
		if err == io.EOF {
			err = ErrEOF
		}
		return string(buf), err
	}

	// If reaches here and is not a file,
	// something was tampered or wrong
	ifile, ok := t.Reader.(*os.File)
	if !ok {
		// Error?
	}

	t.width, _ = t.GetSize() // Update width actually

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
	t.out.WriteString(t.prompt.DisplayString())
	t.Flush()

	//
	rinput := ansi.NewScanner(reader)

	for {
		t.width, _ = t.GetSize()  // Update width actually
		val, err := rinput.Scan() // Ansi scanner
		if err != nil {
			return "", err
		}
		// Do some keyMapping, (i.e: MoveNext: "\t")
		// Plugable Modes like the readline
		// transform val to key
		ch := mapKey(val)

		if t.AutoComplete != nil { // Perform complete on enter
			newLine, newPos, ok := t.AutoComplete(t.prompt.InputString(), t.prompt.Cursor(), ch)
			if ok {
				t.prompt.SetInput(newLine) // Reset print here?
				t.prompt.SetCursor(newPos) // Do nothing
				t.out.WriteString(t.prompt.DisplayString())
				t.Flush()
				//log.Printf("D: %#v", ch)
				continue
				//break
			}
		}

		// Do handle Key operation here
		// Select handler here
		//
		// Non clear
		switch ch {
		case '\x15': // CtrlU
			t.Write([]byte("\033[H"))
		case AKClearScreen: // CtrlL
			ab := ansi.NewHelperDirect(&t.out)
			ab.ClearScreen()
			ab.Move(0, 0)
			//ab.WriteString("\033[2J\033[H") // Clear and save?
			ab.SaveCursor()
			t.Flush()
		case AKeot: //EOT // CtrlD
			if len(t.prompt.InputString()) == 0 {
				return "", ErrEOF
			}
			t.prompt.Delete()

		case AKInterrupt: // CtrlC
			return "", ErrInterrupt // Interrupt return
		// Sequence
		case AKEnter, '\n': // ENTER COMPLETE enter // Process input
			//t.prompt.CursorToEnd() // Index to end of prompt what if?
			t.prompt.extraLine = 0
			t.out.WriteString("\n") // Line feed directly
			//ansi.NewHelperDirect(&t.out).SaveCursor() // ansi for just save cursor?
			t.Flush() // Send to terminal
			if t.prompt.Len() == 0 {
				break
			}
			// Append history
			t.History.Append(t.prompt.InputString())

			line := t.prompt.InputString()
			t.prompt.Clear()
			return line, nil
		case AKWordForward: // Word back
			t.prompt.CursorWordForward()
		case AKWordBack: // Word back
			t.prompt.CursorWordBack()
		case AKDelete: // Escape key delete Weird keys
			t.prompt.Delete()
		case AKBackspace:
			t.prompt.Backspace()
		case AKHistPrev:
			t.prompt.SetInput(t.History.Prev(t.prompt.InputString()))
		case AKHistNext:
			t.prompt.SetInput(t.History.Next(t.prompt.InputString()))
		case AKCursLeft:
			t.prompt.CursorLeft()
		case AKCursRight:
			t.prompt.CursorRight()
		default: // Function that actually adds the text
			if val.Type != ansi.TypeRune {
				break
			}
			//ch, _ := utf8.DecodeRuneInString(val.Value)
			// Go through auto complete
			complete := false
			if !complete && unicode.IsPrint(ch) { // Do not add the tab
				t.prompt.WriteString(val.Value)
			}
		}

		// Auto completer
		/*if t.AutoComplete != nil {
			newLine, newPos, ok := t.AutoComplete(t.prompt.InputString(), t.prompt.Cursor(), ch)
			if ok {
				t.prompt.SetInput(newLine) // Reset print here?
				t.prompt.SetCursor(newPos)
				//break
			}
		}*/

		// Lock here
		t.out.WriteString(t.prompt.DisplayString())
		t.Flush()
	}
}

//Write can be called by any other thread and a Flush triggered
func (t *Term) Write(b []byte) (n int, err error) {
	if len(b) == 0 {
		return 0, nil
	}
	t.prompt.extraLine = 0
	if b[len(b)-1] != '\n' { // Prompt will show below the last cursor in next print
		t.prompt.extraLine = 1
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

	ab.WriteString(t.prompt.DisplayString())
	ab.Flush()
	t.Flush()

	return
}

//GetSize get terminal size returns 1,1 (minimal) on error
func (t *Term) GetSize() (w int, h int) {
	f, ok := t.Writer.(*os.File)
	if !ok {
		return 80, 25 // default unknown size
	}

	iw, ih, err := termutils.GetSize(int(f.Fd()))
	if err != nil {
		return 80, 25
	}

	return iw, ih
}

//SetPrompt sets the terminal prompt
func (t *Term) SetPrompt(p string) {
	t.prompt.SetPrompt(p)
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
	rd := ansi.NewScanner(bytes.NewReader(t.out.Bytes()))

	for {
		val, err := rd.Scan()
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

// Key/Input mapper
func mapKey(val ansi.Value) rune {
	var ch rune
	switch val.Value {
	case "\x0C":
		ch = AKClearScreen
	case "\033f":
		ch = AKWordForward
	case "\033b":
		ch = AKWordBack
	case "\b", "\x7f":
		ch = AKBackspace
	case "\033[A", "\x10": // Up or CtrlP
		ch = AKHistPrev
	case "\033[B", "\x0E": // Down or CtrlN
		ch = AKHistNext
	case "\033[D", "\x02": // left or CtrlB
		ch = AKCursLeft
	case "\033[C", "\x06": // Right or ctrlF
		ch = AKCursRight
	case "\033[Z": // Right or ctrlF
		ch = AKShiftTab
	case "\033[3~":
		ch = AKDelete
	default:
		ch, _ = utf8.DecodeRuneInString(val.Value)
	}
	return ch

}
