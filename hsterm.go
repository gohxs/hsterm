// hsterm terminal thing with history, complete (not done yet)
// needs code cleanup, maybe stop the clear string instead of redraw

package hsterm

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	glog "log"
	"os"
	"unicode"

	"github.com/gohxs/hsterm/internal/ansireader"
	"github.com/gohxs/hsterm/internal/term"
	"github.com/gohxs/prettylog"
)

var (
	ErrInterrupt = errors.New("Interrupt")
	log          = prettylog.New("hsterm", Stdout)
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

	dispbuf bytes.Buffer

	tstate *term.State
	inFile io.Reader
	//inFile *os.File // io.Reader
	io.Writer
}

//New instantiates a new Terminal handler
func New() *Term {

	inFile := Stdin
	outFile := Stdout

	// For windows only
	//inFD := syscall.STD_INPUT_HANDLE
	//outFD := syscall.STD_OUTPUT_HANDLE

	state, _ := term.MakeRaw(int(term.GetStdin()))
	width := term.GetScreenWidth()

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
		inFile:       inFile,
		Writer:       outFile, // TODO: for now
	}
	f, err := os.OpenFile("out.txt", os.O_APPEND|os.O_WRONLY, os.FileMode(0644))
	if err != nil {
		panic(err)
	}
	log = prettylog.New("Term", &WriteFlush{f})
	return ret
}

// Close and restore terminal state
func (t *Term) Close() {
	term.Restore(int(term.GetStdin()), t.tstate)
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
	/*if !term.IsTerminal(int(Stdin.Fd())) {
		rd := bufio.NewScanner(frd)
		for rd.Scan() {
			curString := rd.Text()
			return curString, nil
		}
	}*/

	// TODO: Wrong should go back to previous position, because on wrap it breaks
	// the output
	//wr.Write([]byte("\r" + t.prompt)) // restore cursor and print prompt?
	// Read command etcetc
	//tin := make([]byte, 128)
	rinput := ansireader.New(frd)
	//rd := bufio.NewReader(frd)
	for {
		val, err := rinput.Read()
		if err != nil {
			//log.Println("Error ocurred reading fd", err, " ilen", len(input))
			break
		}

		switch val.Ch {
		case ansireader.CharEOT:
			t.ClearDisplay() // Debug
			continue
		case ansireader.CharInterrupt:
			return "", ErrInterrupt // Interrupt return
		case ansireader.CharCtrlJ, ansireader.CharEnter: // Carriage return followed by \n i supose
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
		switch val.Ch {
		case ansireader.CharCtrlL:
			t.Writer.Write([]byte("\033[2J\033[H")) // Clear
		case ansireader.CharDelete: // Escape key delete
			t.inbuf.Delete()
		case ansireader.CharBackspace, 8:
			log.Println("Backspace")
			t.inbuf.Backspace()
		case ansireader.CharPrev:
			t.histindex--
			if t.histindex < 0 {
				t.histindex = 0
			} else {
				t.inbuf.Set(t.history[t.histindex])
			}
		case ansireader.CharNext:
			t.histindex++
			if t.histindex >= len(t.history) { // Clear input if above the len
				t.histindex = len(t.history)
				t.inbuf.Clear()
			} else {
				t.inbuf.Set(t.history[t.histindex])
			}
		case ansireader.CharBackward:
			t.inbuf.CursorLeft()
		case ansireader.CharForward:
			t.inbuf.CursorRight()
		default: // Function that actually adds the text
			// Go through auto complete
			if t.AutoComplete != nil {
				newLine, newPos, ok := t.AutoComplete(t.inbuf.String(), t.inbuf.Cursor(), val.Ch)
				if ok {
					t.inbuf.Set(newLine) // Reset print here?
					log.Println("NewPos:", newPos)
					t.inbuf.SetCursor(newPos)
					break
				}
			}
			if !unicode.IsPrint(val.Ch) { // Unhandled char
				log.Printf("b[%d] %v\n", 1, val.Ch) // Print unprintable char
				break
			}
			t.inbuf.WriteRune(val.Ch)
		}
		t.RefreshDisplay()
	}
	return "", nil
}

//ClearDisplay clears the prompt display
func (t *Term) ClearDisplay() { // OrClear?
	var (
		lWidth  = t.width
		lPrompt = len(t.prompt)
		lBuf    = t.inbuf.Len() // Less last added char i supose?
	)
	lineLen := lBuf + lPrompt
	count := (lineLen) / lWidth

	t.dispbuf.Reset()
	if count > 0 {
		//t.dispbuf.WriteString(strings.Repeat("\033[2K\r\033[A", count)) // clear and moveup?
		//t.dispbuf.WriteString("\033[2K")                                // clear and moveup?
		t.dispbuf.Write([]byte(fmt.Sprintf("\033[%dA\r\033[J", count))) // clear and moveup?
	} else {
		t.dispbuf.Write([]byte(fmt.Sprintf("\r\033[J"))) // clear and moveup?
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
	// Calculate cursor index for print

	//TODO: Disable cursor, print move enable enable
	t.dispbuf.WriteString(fmt.Sprintf("\r%s%s \b", t.prompt, dispBuf)) // Erase after?

	// Position cursor
	// Cur left and up if any? of current line?
	// Disable this for now?
	dispLen := t.inbuf.Len()
	curLeft := (dispLen - t.inbuf.Cursor())
	if curLeft != 0 && dispLen > 0 {
		t.dispbuf.WriteString(fmt.Sprintf("\033[%dD", curLeft)) // ? huh
	}

	t.Writer.Write(t.dispbuf.Bytes()) // flush
	t.dispbuf.Reset()

}
