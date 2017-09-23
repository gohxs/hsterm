package ansireader

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

var (
	defValue = Value{}
)

type Value struct { // We can add extra values in case of ascii parse
	// Value Type
	Type int
	// Int value to handle \033[{number};{number}code
	Attr []int
	// Key value
	Ch rune
}

// Reader translated reader of terminal
type Reader struct {
	breader *bufio.Reader
	//rd      io.Reader
}

func New(rd io.Reader) *Reader {
	return &Reader{bufio.NewReader(rd)}
}

// ReadOne Ansi char
func (r Reader) Read() (Value, error) {
	ret := Value{}
	ch, _, err := r.breader.ReadRune()
	if err != nil {
		return Value{}, err //log.Println("Error ocurred reading fd", err, " ilen", len(input))
	}
	if ch == CharEsc {
		ret.Type = TypeMeta
		ch, _, _ = r.breader.ReadRune() // NextRune
		if ch == '[' {                  // Escape key ex?
			ch, _, _ = r.breader.ReadRune() // Or byte?
			return r.readEscKey(ch)
		}
		return Value{TypeMeta, nil, ch}, nil
		//return r.escapeKey(ch), nil // Meta?
	}

	// Normal char
	return Value{TypeRune, nil, ch}, nil
}

func (r Reader) readEscKey(ch rune) (val Value, err error) {
	attrbuf := bytes.NewBuffer(nil) // attrbuf

	if ch == '?' { // next  for cases like \033[? // vt100
		ch, _, err = r.breader.ReadRune()
	}

	for ch == ';' || unicode.IsNumber(ch) {
		attrbuf.WriteRune(ch)
		ch, _, _ = r.breader.ReadRune()
	}
	val.Ch = ch
	// Parse attr
	attr := attrbuf.String()
	var param []int
	// Parse escAttr.attr and put to Param
	par := strings.Split(attr, ";") // In some cases ':' 16m color
	for _, pv := range par {
		iv, _ := strconv.Atoi(pv)
		param = append(param, iv)
	}
	val.Attr = param

	return
}

// Partial code from chzyer/readline
//
// translate Esc[X
/*func (r Reader) escapeExKey(key *escapeKeyPair) rune {
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
}*/

// translate EscX to Meta+X
/*func (r Reader) escapeKey(ch rune) Value {

	if ch == CharEscapeEx { // Readmore?
		log.Println("Esc is ex:", ch)

		echar, _, _ := r.breader.ReadRune()
		ex := r.readEscKey(echar)
		return Value{Ch: r.escapeExKey(ex)}
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
	return Value{Type: TypeMeta, Attr: nil, Ch: ch}
}*/
