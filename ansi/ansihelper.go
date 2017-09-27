// Wrong should be different, this is a helper

package ansi

import (
	"fmt"
	"io"
)

// Maybe const this

// Writer struct that holds methods to easy write ansi escape codes
type Helper struct {
	io.Writer
}

func NewHelper(wr io.Writer) *Helper {
	return &Helper{wr}
}

func (w *Helper) WriteString(s string) (int, error) {
	return w.Writer.Write([]byte(s))
}

func (w *Helper) SaveCursor() {
	w.WriteString("\033[s")
}

//WriteRestoreCursor writes the escape for restoring cursor
func (w *Helper) RestoreCursor() {
	w.WriteString("\033[u")
}

// Cursor movement
func (w *Helper) MoveUp(n int) {
	if n < 1 {
		return
	}
	w.WriteString(fmt.Sprintf("\033[%dA", n))
}
func (w *Helper) MoveDown(n int) {
	if n < 1 {
		return
	}
	w.WriteString(fmt.Sprintf("\033[%dB", n))
}
func (w *Helper) MoveLeft(n int) {
	if n < 1 {
		return
	}
	w.WriteString(fmt.Sprintf("\033[%dD", n))
}
func (w *Helper) MoveRight(n int) {
	if n < 1 {
		return
	}
	w.WriteString(fmt.Sprintf("\033[%dC", n))
}
