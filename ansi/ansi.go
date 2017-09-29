package ansi

import (
	"bytes"
	"strings"
)

var (
	defValue = Value{}
)

// AnsiParser // reader
const (
	TypeRune = iota
	TypeMeta
	TypeEscape
	TypeString // Not used yet but can be used to return a fullstring(nonescape) in a value
)

// Value Ansi Escape sequence value
type Value struct { // We can add extra values in case of ascii parse
	Raw string
	// Value Type
	Type int
	// Int value to handle \033[{number};{number}code
	Attr []int
	// Key value
	Value string
}

//Strip ansi from string
func Strip(in string) string {
	buf := bytes.NewBuffer(nil)
	scanner := NewScanner(strings.NewReader(in))
	for {
		val, err := scanner.Scan()
		if err != nil {
			break
		}
		if val.Type == TypeRune {
			buf.WriteString(val.Value)
		}
	}
	return buf.String()
}
