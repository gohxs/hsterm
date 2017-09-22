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
	"unicode"

	"github.com/gohxs/hsterm/internal/term"
	"github.com/gohxs/prettylog"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	ErrInterrupt = errors.New("Interrupt")
	log          = prettylog.New("hsterm")
	flog         *glog.Logger
)

const (
	chCtrlC     = 3
	chBackspace = 8
	chDelete    = 127
	chTab       = 9
	chEsc       = 27
)

// TermFunc callback for receiving a command
//type termFunc func(string)

// Term main var
type Term struct {
	inbuf *InputBuffer
	//rbuf      string
	//cursIndex int
	prompt string

	width int

	// History class somewhere
	histindex int
	history   []string

	// Callbacks
	Logger       *glog.Logger
	Display      func(string) string
	AutoComplete func(line string, pos int, key rune) (newLine string, newPos int, ok bool)

	tstate *term.State
	inFile *os.File // io.Reader
	io.Writer
}

//New instantiates a new Terminal handler
func New() *Term {
	frd := os.Stdin

	state, _ := term.MakeRaw(int(frd.Fd()))

	width, _, err := term.GetSize(int(frd.Fd()))
	log.Println("Terminal is:", width, "Wide")

	ret := &Term{
		inbuf:        NewInputBuffer(),
		prompt:       "",
		width:        width,
		histindex:    0,
		history:      []string{},
		Logger:       prettylog.Dummy(),
		Display:      nil,
		AutoComplete: nil,
		tstate:       state,
		inFile:       frd,
		Writer:       os.Stdout, // TODO: for now
	}
	f, err := os.OpenFile("out.txt", os.O_APPEND|os.O_WRONLY, os.FileMode(0644))
	if err != nil {
		panic(err)
	}
	log = prettylog.New("Term", &WriteFlush{f})
	return ret
}

func (t *Term) Close() {
	term.Restore(int(t.inFile.Fd()), t.tstate)
}

func (t *Term) Write(b []byte) (n int, err error) {
	buf := bytes.NewBuffer(nil)
	buf.WriteString("\r\033[2K") // Go back? and clear rest of line
	// Go up and back, print, and comeback
	n, err = buf.Write(b) // Print content
	t.Writer.Write(buf.Bytes())
	t.RefreshDisplay() // do not reset position
	return
}

// Prompt sets the terminal prompt
func (t *Term) Prompt(p string) {
	t.prompt = p
}

// Readline readline readline
func (t *Term) Readline() (string, error) {
	//wr := os.Stdout
	frd := t.inFile

	//defer term.Restore(int(t.inFile.Fd()), state)

	// Some way to set this
	if !terminal.IsTerminal(int(frd.Fd())) {
		rd := bufio.NewScanner(frd)
		for rd.Scan() {
			curString := rd.Text()
			return curString, nil
		}
	}

	// TODO: Wrong should go back to previous position, because on wrap it breaks
	// the output
	//wr.Write([]byte("\r" + t.prompt)) // restore cursor and print prompt?
	// Read command etcetc
	//tin := make([]byte, 128)
	rinput := NewReader(frd)
	//rd := bufio.NewReader(frd)
	for {
		val, err := rinput.Read()
		if err != nil {
			//log.Println("Error ocurred reading fd", err, " ilen", len(input))
			break
		}

		switch val.ch {
		case CharEOT:
			t.ClearDisplay() // Debug
			continue
		case CharInterrupt:
			return "", ErrInterrupt // Interrupt return
		case CharCtrlJ, CharEnter: // Carriage return followed by \n i supose
			//t.ResetCursor()
			t.Writer.Write([]byte("\n")) // output
			if t.inbuf.Len() != 0 {      // process
				t.history = append(t.history, t.inbuf.String())
				t.histindex = len(t.history)

				line := t.inbuf.String()
				t.inbuf.Clear()

				t.RefreshDisplay() // Output the prompt etc

				return line, nil
				// Process callbacks
			}
			t.RefreshDisplay()
		}

		t.ClearDisplay()
		switch val.ch {
		case CharCtrlL:
			t.Writer.Write([]byte("\033[2J\033[H")) // Clear
		case CharDelete: // Escape key delete
			t.inbuf.Delete()
		case CharBackspace:
			t.inbuf.Backspace()
		case CharPrev:
			t.histindex--
			if t.histindex < 0 {
				t.histindex = 0
			} else {
				t.inbuf.Set(t.history[t.histindex])
			}
		case CharNext:
			t.histindex++
			if t.histindex >= len(t.history) { // Clear input if above the len
				t.histindex = len(t.history)
				t.inbuf.Clear()
			} else {
				t.inbuf.Set(t.history[t.histindex])
			}
		case CharBackward:
			t.inbuf.CursorLeft()
		case CharForward:
			t.inbuf.CursorRight()
		default: // Function that actually adds the text
			// Go through auto complete
			if t.AutoComplete != nil {
				newLine, newPos, ok := t.AutoComplete(t.inbuf.String(), t.inbuf.Cursor(), val.ch)
				if ok {
					t.inbuf.Set(newLine) // Reset print here?
					log.Println("NewPos:", newPos)
					t.inbuf.SetCursor(newPos)
					break
				}
			}
			if !unicode.IsPrint(val.ch) { // Unhandled char
				log.Printf("b[%d] %v\n", 1, val.ch) // Print unprintable char
				break
			}
			t.inbuf.WriteRune(val.ch)
		}
		t.RefreshDisplay()
	}
	return "", nil
}

//Clears the prompt display
func (t *Term) ClearDisplay() { // OrClear?
	var (
		lWidth  = t.width
		lPrompt = len(t.prompt)
		lBuf    = t.inbuf.Len() // Less last added char i supose?
	)
	lineLen := lBuf + lPrompt
	count := (lineLen) / lWidth
	if count > 0 {
		//t.Writer.Write([]byte(strings.Repeat("\r\033[K\033[A\033[K", count))) // clear and moveup?
		t.Writer.Write([]byte(fmt.Sprintf("\033[%dA\r\033[J", count))) // clear and moveup?
	} else {
		t.Writer.Write([]byte(fmt.Sprintf("\r\033[J"))) // clear and moveup?
	}
}

// RefreshDisplay the prompt and input
func (t *Term) RefreshDisplay() {
	// TODO: Wrong should go back to previous position, because on wrap it breaks
	// the output
	// Transform output
	var dispBuf string
	if t.Display != nil {
		dispBuf = t.Display(t.inbuf.String())
	} else {
		dispBuf = t.inbuf.String()
	}
	out := bytes.NewBuffer(nil)
	// Calculate cursor index for print

	out.WriteString(fmt.Sprintf("\r%s%s \b\033[K", t.prompt, dispBuf))

	// Position cursor
	// Cur left and up if any? of current line?
	dispLen := t.inbuf.Len()
	curLeft := (dispLen - t.inbuf.Cursor())
	//log.Println("CurLeft:", curLeft, dispLen, t.inbuf.Cursor())
	// Move cursor to where it belong??
	// After print we can reset Pos, move back and forward
	if curLeft != 0 && dispLen > 0 {
		out.WriteString(fmt.Sprintf("\033[%dD", curLeft)) // ? huh
	}
	t.Writer.Write(out.Bytes())

}
