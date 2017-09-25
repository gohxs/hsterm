package ansi

var (
	defValue = Value{}
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

// AnsiParser // reader
const (
	TypeRune = iota
	TypeMeta
	TypeEscape
)
