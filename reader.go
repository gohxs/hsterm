package hsterm

////////////////////
// this is suposed to be an ansi reader
// which returns an array of Ansi values
//
//
import (
	"bufio"
	"bytes"
	"io"
	"strconv"
	"strings"
	"unicode"
)

// AnsiParser // reader

type Value struct { // We can add extra values in case of ascii parse
	// Escaped params
	ch rune
}

var (
	defValue = Value{}
)

// Reader translated reader of terminal
type Reader struct {
	breader *bufio.Reader
	//rd      io.Reader
}

func NewReader(rd io.Reader) *Reader {
	return &Reader{bufio.NewReader(rd)}
}

// ReadOne
func (r Reader) Read() (Value, error) {
	ch, _, err := r.breader.ReadRune()
	if err != nil {
		return Value{}, err //log.Println("Error ocurred reading fd", err, " ilen", len(input))
	}
	//input = append(input, tin[:r]...)
	//log.Println("Debug input:", ch)
	if ch == CharEsc {
		//log.Println("Char is esc, Reading next")
		echar, _, _ := r.breader.ReadRune() // NextRune
		return r.escapeKey(echar), nil
	}
	return Value{ch: ch}, nil
}

// Partial code from chzyer/readline
//
// translate Esc[X
func (r Reader) escapeExKey(key *escapeKeyPair) rune {
	var ch rune
	switch key.typ {
	case 'Z':
		ch = MetaShiftTab
	case 'D':
		ch = CharBackward
	case 'C':
		ch = CharForward
	case 'A':
		ch = CharPrev
	case 'B':
		ch = CharNext
	case 'H':
		ch = CharLineStart
	case 'F':
		ch = CharLineEnd
	case '~':
		switch key.attr {
		case "3":
			ch = CharDelete
		case "1":
			ch = CharLineStart
		case "4":
			ch = CharLineEnd
		}
	default:
	}
	return ch
}

// Better escape thing
type escapeKeyPair struct {
	attr string
	typ  rune
}

// Why?
func (e *escapeKeyPair) Get2() (int, int, bool) {
	sp := strings.Split(e.attr, ";")
	if len(sp) < 2 {
		return -1, -1, false
	}
	s1, err := strconv.Atoi(sp[0])
	if err != nil {
		return -1, -1, false
	}
	s2, err := strconv.Atoi(sp[1])
	if err != nil {
		return -1, -1, false
	}
	return s1, s2, true
}

func (r Reader) readEscKey(ch rune) *escapeKeyPair {
	p := escapeKeyPair{}
	buf := bytes.NewBuffer(nil)
	for {
		if ch == ';' {
		} else if unicode.IsNumber(ch) {
		} else {
			p.typ = ch
			break
		}
		buf.WriteRune(ch)
		ch, _, _ = r.breader.ReadRune()
	}
	p.attr = buf.String()
	return &p
}

// translate EscX to Meta+X
func (r Reader) escapeKey(ch rune) Value {

	if ch == CharEscapeEx { // Readmore?
		log.Println("Esc is ex:", ch)

		echar, _, _ := r.breader.ReadRune()
		ex := r.readEscKey(echar)
		return Value{ch: r.escapeExKey(ex)}
	}

	switch ch {
	case 'b':
		ch = MetaBackward
	case 'f':
		ch = MetaForward
	case 'd':
		ch = MetaDelete
	case CharTranspose:
		ch = MetaTranspose
	case CharBackspace:
		ch = MetaBackspace
	case 'O':
		d, _, _ := r.breader.ReadRune()
		switch d {
		case 'H':
			ch = CharLineStart
		case 'F':
			ch = CharLineEnd
		default:
			r.breader.UnreadRune()
		}
	case CharEsc:

	}
	return Value{ch: ch}
}

const (
	CharEOF       = 0 // EOF
	CharLineStart = 1
	CharBackward  = 2
	CharInterrupt = 3
	CharEOT       = 4 // EOF? // End of Transmition
	CharLineEnd   = 5
	CharForward   = 6
	CharBell      = 7
	CharCtrlH     = 8
	CharTab       = 9
	CharCtrlJ     = 10
	CharKill      = 11
	CharCtrlL     = 12
	CharEnter     = 13
	CharNext      = 14
	CharPrev      = 16
	CharBckSearch = 18
	CharFwdSearch = 19
	CharTranspose = 20
	CharCtrlU     = 21
	CharCtrlW     = 23
	CharCtrlZ     = 26
	CharEsc       = 27
	CharEscapeEx  = 91
	CharBackspace = 127
)

const (
	MetaBackward rune = -iota - 1
	MetaForward
	MetaDelete
	MetaBackspace
	MetaTranspose
	MetaShiftTab
	CharDelete
)
