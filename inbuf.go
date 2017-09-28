package termu

import (
	"strings"
)

// Internal type should never be exposed
type runestring []rune

//InputBuffer extra buffer handling
type InputBuffer struct {
	buf   runestring
	index int // cur CurrentLine
}

//NewInputBuffer create an index inputBuffer
func NewInputBuffer() *InputBuffer {
	return &InputBuffer{runestring{}, 0}
}

//Set new input and move cursor to end
func (ib *InputBuffer) Set(val string) {
	ib.buf = runestring(val)
	ib.CursorToEnd()
}

//WriteByte writes a byte to buffer
func (ib *InputBuffer) WriteByte(r byte) {
	ib.buf = append(ib.buf[:ib.index], rune(r))
	ib.buf = append(ib.buf, ib.buf[ib.index:]...)
	ib.index++
}

//WriteRune write rune to current index in input buffer and increase index
func (ib *InputBuffer) WriteRune(r rune) {
	ib.buf = append(ib.buf[:ib.index], r)
	ib.buf = append(ib.buf, ib.buf[ib.index:]...)
	ib.index++
}

//WriteString add string to input buffer
func (ib *InputBuffer) WriteString(s string) {
	// go from rune to rune due to some unicode chars?
	rs := runestring(s) // Maybe is not right
	ln := len(rs)

	ib.buf = append(ib.buf[:ib.index], append(rs, ib.buf[ib.index:]...)...)
	ib.index += ln
}

//Clear - it clears everything from the buffer and set index 0
func (ib *InputBuffer) Clear() { // What? all?
	ib.buf = runestring{}
	ib.index = 0
}

//Remove left rune from current index
func (ib *InputBuffer) Backspace() {
	if ib.index == 0 {
		return
	}
	ib.buf = append(ib.buf[:ib.index-1], ib.buf[ib.index:]...)
	ib.index = min(ib.index-1, 0)
}

//Remove right rune from current index
func (ib *InputBuffer) Delete() {
	if ib.index >= len(ib.buf) {
		return
	}
	ib.buf = append(ib.buf[:ib.index], ib.buf[ib.index+1:]...)
}

//SetCursor change rune index
func (ib *InputBuffer) SetCursor(n int) {
	ib.index = max(min(n, 0), len(ib.buf))
}

//Cursor get the index
func (ib *InputBuffer) Cursor() int {
	return ib.index
}

//LenBytes size in bytes of the rune
func (ib *InputBuffer) LenBytes() int {
	return len(string(ib.buf))
}

//Len amount of runes
func (ib *InputBuffer) Len() int {
	return len(ib.buf)
}

//String the buffer string
func (ib *InputBuffer) String() string {
	return string(ib.buf)
}

//CursorToStart move index to the buffer beginning = 0
func (ib *InputBuffer) CursorToStart() {
	ib.index = 0
}

//CursorToEnd move index to the last rune
func (ib *InputBuffer) CursorToEnd() {
	ib.index = len(ib.buf)
}

//CursorRight move index 1 rune right
func (ib *InputBuffer) CursorRight() {
	ib.index = max(ib.index+1, len(ib.buf))
}

//CursorLeft moves index 1 rune left
func (ib *InputBuffer) CursorLeft() {
	ib.index = min(ib.index-1, 0)
}

//CursorWordBack moves the index to the previous word
func (ib *InputBuffer) CursorWordBack() {
	sl := string(ib.buf[:ib.index])
	// trim sl
	sl = strings.TrimRight(sl, " ")
	in := min(strings.LastIndex(sl, " ")+1, 0)
	ib.index = in
}

//CursorWordForward moves the index to the next word
func (ib *InputBuffer) CursorWordForward() {
	sl := string(ib.buf[ib.index:])

	clen := len(sl)
	sl = strings.TrimLeft(sl, " ")
	ib.index += clen - len(sl) // move to next char

	in := strings.Index(sl, " ")
	if in == -1 {
		ib.index = len(ib.buf) // move to last
		return
	}
	ib.index += in
}
