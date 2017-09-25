package hsterm

import (
	"strings"
)

// Internal type should never be exposed
type runestring []rune

//InputBuffer extra buffer handling
type InputBuffer struct {
	//buf       []string // multiple struct // Multiline
	buf       runestring
	curLine   int
	cursIndex int // cur CurrentLine
}

func NewInputBuffer() *InputBuffer {
	return &InputBuffer{runestring{}, 0, 0}
}

func (ib *InputBuffer) Set(val string) {
	ib.buf = runestring(val)
	ib.CursorToEnd()
}
func (ib *InputBuffer) WriteByte(r byte) {
	//	line := ib.buf[ib.curLine]
	//	line = line[:ib.cursIndex] + string(r) + line[ib.cursIndex:]
	//	ib.buf[ib.curLine] = line

	ib.buf = append(ib.buf[:ib.cursIndex], runestring(string(r))...)
	ib.buf = append(ib.buf, ib.buf[ib.cursIndex:]...)
	//ib.buf[:ib.cursIndex] + runestring(string(r)) + ib.buf[ib.cursIndex:]

	ib.cursIndex++
}

//WriteRune  Add rune to input buffer
func (ib *InputBuffer) WriteRune(r rune) {
	//line := ib.buf[ib.curLine]
	//line = line[:ib.cursIndex] + string(r) + line[ib.cursIndex:]
	//ib.buf[ib.curLine] = line
	ib.buf = append(ib.buf, r)
	ib.cursIndex++
	//ib.cursIndex += utf8.RuneLen(r) // Wrong
}

//WriteString add string to input buffer
func (ib *InputBuffer) WriteString(s string) {
	//line := ib.buf[ib.curLine]
	//line = line[:ib.cursIndex] + s + line[ib.cursIndex:]
	//ib.buf[ib.curLine] = line
	//
	for _, r := range s { // go from rune to rune
		ib.buf = append(ib.buf, r)
		ib.cursIndex++ // len(s)
	}
}

//Clear - it clears everything from the buffer
func (ib *InputBuffer) Clear() { // What? all?
	ib.buf = runestring{}
	ib.curLine = 0
	ib.cursIndex = 0
}

// Remove last rune not last byte
func (ib *InputBuffer) Backspace() {
	if ib.cursIndex == 0 {
		return
	}
	//line := ib.buf[ib.curLine]
	//line = line[:ib.cursIndex-1] + line[ib.cursIndex:]
	//ib.buf[ib.curLine] = line
	ib.buf = append(ib.buf[:ib.cursIndex-1], ib.buf[ib.cursIndex:]...)
	ib.cursIndex = min(ib.cursIndex-1, 0)
}

func (ib *InputBuffer) Delete() {
	if ib.cursIndex >= len(ib.buf) { // ignroe?
		return
	}
	//line := ib.buf[ib.curLine]
	//line = line[:ib.cursIndex] + line[ib.cursIndex+1:]
	//ib.buf[ib.curLine] = line
	ib.buf = append(ib.buf[:ib.cursIndex], ib.buf[ib.cursIndex+1:]...)
}

func (ib *InputBuffer) SetCursor(n int) {
	ib.cursIndex = max(min(n, 0), len(ib.buf))
}
func (ib *InputBuffer) Cursor() int {
	return ib.cursIndex
}

func (ib *InputBuffer) LenBytes() int {
	return len(string(ib.buf))
}
func (ib *InputBuffer) Len() int {
	return len(ib.buf)
}

// All?
func (ib *InputBuffer) String() string {
	return string(ib.buf)
}

func (ib *InputBuffer) CursorToStart() {
	ib.cursIndex = 0
}
func (ib *InputBuffer) CursorToEnd() {
	ib.cursIndex = len(ib.buf)
}
func (ib *InputBuffer) CursorRight() {
	ib.cursIndex = max(ib.cursIndex+1, len(ib.buf))
}
func (ib *InputBuffer) CursorLeft() {
	ib.cursIndex = min(ib.cursIndex-1, 0)
}
func (ib *InputBuffer) CursorWordBack() {
	sl := string(ib.buf[:ib.cursIndex])
	// trim sl
	sl = strings.TrimRight(sl, " ")
	in := min(strings.LastIndex(sl, " ")+1, 0)
	ib.cursIndex = in
}

func (ib *InputBuffer) CursorWordForward() {
	sl := string(ib.buf[ib.cursIndex:])

	clen := len(sl)
	sl = strings.TrimLeft(sl, " ")
	ib.cursIndex += clen - len(sl) // move to next char

	in := strings.Index(sl, " ")
	if in == -1 {
		ib.cursIndex = len(ib.buf) // move to last
		return
	}
	ib.cursIndex += in
}
