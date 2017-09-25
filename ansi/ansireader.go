package ansi

////////////////////
// this is suposed to be an ansi reader
// which returns an array of Ansi values
import (
	"bufio"
	"bytes"
	"io"
	"strconv"
	"strings"
	"unicode"
)

// Reader translated reader of terminal
type Reader struct {
	Reader *bufio.Reader // ??
}

// NewReader create a reader
func NewReader(rd io.Reader) *Reader {
	return &Reader{bufio.NewReader(rd)}
}

//Read reads stripping Ansi codes
func (r Reader) Read(b []byte) (int, error) {
	// Read a bunch from reader,
	// Create sub Reader, and return thing
	tb := make([]byte, 1024)    // 1k buffer for now
	n, err := r.Reader.Read(tb) // Raw read
	// ignore err here?
	ar := NewReader(bytes.NewReader(tb[:n]))

	readed := 0 // Parse one by one from all readed
	for {
		val, loerr := ar.ReadEscape()
		if loerr != nil {
			return readed, err
		}

		if val.Type != TypeRune { // Skip escapes
			continue
		}
		if readed < len(b) {
			copy(b[readed:], val.Value)
		}
		readed += len([]byte(val.Value))
	}
}

// ReadEscape reads a value from reader, either escape or char
func (r Reader) ReadEscape() (Value, error) {
	// Create a writter instead of adding all
	ret := Value{}
	ch, _, err := r.Reader.ReadRune()
	if err != nil {
		return Value{}, err //log.Println("Error ocurred reading fd", err, " ilen", len(input))
	}
	ret.Value += string(ch)
	ret.Raw += string(ch)

	if ch == '\033' {
		ret.Type = TypeMeta
		ch, _, _ = r.Reader.ReadRune() // NextRune
		ret.Value += string(ch)
		ret.Raw += string(ch)
		switch ch {
		case 'O': // O key requires another
			ch, _, _ = r.Reader.ReadRune() // Or byte?
			ret.Value += string(ch)
			ret.Raw += string(ch)
		case '[':
			ch, _, _ = r.Reader.ReadRune() // Or byte?
			ret.Raw += string(ch)
			r.readSeq(ch, &ret)
		}
	}
	// Normal char
	return ret, nil
}

func (r Reader) readSeq(ch rune, val *Value) (err error) {
	attrbuf := bytes.NewBuffer(nil) // attrbuf
	val.Type = TypeEscape

	noAttr := false

	if ch == '?' || ch == '>' { // prefixed CSI
		val.Value += string(ch)          // Add the thing
		ch, _, err = r.Reader.ReadRune() // Next
		val.Raw += string(ch)
	}

	// Read and parse Params
	for ch == ';' || unicode.IsNumber(ch) {
		attrbuf.WriteRune(ch)
		ch, _, _ = r.Reader.ReadRune()
		val.Raw += string(ch)
	}
	if ch == '~' { // ends with ~
		noAttr = true
	}
	if noAttr {
		val.Value += attrbuf.String() + string(ch)
		return
	}

	// Parse attr
	attr := attrbuf.String()
	if len(attr) > 0 {
		var param []int
		// Parse escAttr.attr and put to Param
		par := strings.Split(attr, ";") // In some cases ':' 16m color
		for _, pv := range par {
			iv, _ := strconv.Atoi(pv)
			param = append(param, iv)
		}
		val.Attr = param
	}
	val.Value += string(ch) // Add the thing

	switch ch {
	case '*', '+', '\'', '$', '"', '%': // taken from konsole testing
		ch, _, _ = r.Reader.ReadRune()
		val.Raw += string(ch)
		val.Value += string(ch) // Add the thing
	}
	return err
}
