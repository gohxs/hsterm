package hsterm

import "strings"

//InputBuffer extra buffer handling
type InputBuffer struct {
	buf       []string // multiple struct // Multiline
	curLine   int
	cursIndex int // cur CurrentLine
}

func NewInputBuffer() *InputBuffer {
	return &InputBuffer{[]string{""}, 0, 0}
}

func (ib *InputBuffer) Set(val string) {
	ib.buf[ib.curLine] = val
	ib.CursorToEnd()
}
func (ib *InputBuffer) WriteByte(r byte) {
	line := ib.buf[ib.curLine]
	line = line[:ib.cursIndex] + string(r) + line[ib.cursIndex:]
	ib.buf[ib.curLine] = line
	ib.cursIndex++
}

//WriteRune  Add rune to input buffer
func (ib *InputBuffer) WriteRune(r rune) {
	line := ib.buf[ib.curLine]
	line = line[:ib.cursIndex] + string(r) + line[ib.cursIndex:]
	ib.buf[ib.curLine] = line
	ib.cursIndex++
}

//Clear - it clears everything from the buffer
func (ib *InputBuffer) Clear() { // What? all?
	ib.buf = []string{""}
	ib.curLine = 0
	ib.cursIndex = 0
}
func (ib *InputBuffer) Backspace() {
	if ib.cursIndex == 0 {
		return
	}
	line := ib.buf[ib.curLine]
	line = line[:ib.cursIndex-1] + line[ib.cursIndex:]
	ib.buf[ib.curLine] = line
	ib.cursIndex = min(ib.cursIndex-1, 0)
}

func (ib *InputBuffer) Delete() {
	if ib.cursIndex >= len(ib.buf[ib.curLine]) { // ignroe?
		return
	}
	line := ib.buf[ib.curLine]
	line = line[:ib.cursIndex] + line[ib.cursIndex+1:]
	ib.buf[ib.curLine] = line
}

func (ib *InputBuffer) SetCursor(n int) {
	ib.cursIndex = max(min(n, 0), len(ib.buf[ib.curLine]))
}
func (ib *InputBuffer) Cursor() int {
	return ib.cursIndex
}
func (ib *InputBuffer) Len() int {
	return len(ib.buf[ib.curLine])
}

// All?
func (ib *InputBuffer) String() string {
	if len(ib.buf) == 1 {
		return ib.buf[ib.curLine]
	}
	return strings.Join(ib.buf, "\n")
}

func (ib *InputBuffer) CursorToStart() {
	ib.cursIndex = 0
}
func (ib *InputBuffer) CursorToEnd() {
	ib.cursIndex = len(ib.buf[ib.curLine])
}

func (ib *InputBuffer) CursorRight() {
	ib.cursIndex = max(ib.cursIndex+1, len(ib.buf[ib.curLine]))
}
func (ib *InputBuffer) CursorLeft() {
	ib.cursIndex = min(ib.cursIndex-1, 0)
}
func (ib *InputBuffer) CursorWordBack() {
	sl := ib.buf[ib.curLine][:ib.cursIndex]
	// trim sl
	sl = strings.TrimRight(sl, " ")
	in := min(strings.LastIndex(sl, " ")+1, 0)

	ib.cursIndex = in
}
func (ib *InputBuffer) CursorWordForward() {
	sl := ib.buf[ib.curLine][ib.cursIndex:]
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

// Total count??
/*func (ib *InputBuffer) LineWrapCount(char intwidth int) int { // Prompt
	lines := 0
	mcount := 0
	for _, c := range ib.buf {
		if mcount > width {
			lines++
			mcount = 0
		}
		mcount++
	}
	return lines
}*/

// Count lines based on width
//
// Perform operations on index
