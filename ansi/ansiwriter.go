package ansi

import (
	"fmt"
	"io"
)

// Maybe const this

// Writer struct that holds methods to easy write ansi escape codes
type Writer struct {
	io.Writer
}

func NewWriter(wr io.Writer) *Writer {
	return &Writer{wr}
}

func (w *Writer) WriteString(s string) (int, error) {
	return w.Writer.Write([]byte(s))
}

func (w *Writer) SaveCursor() {
	w.WriteString("\033[s")
}

//WriteRestoreCursor writes the escape for restoring cursor
func (w *Writer) RestoreCursor() {
	w.WriteString("\033[u")
}

// Cursor movement
func (w *Writer) MoveUp(n int) {
	if n < 1 {
		return
	}
	w.WriteString(fmt.Sprintf("\033[%dA", n))
}
func (w *Writer) MoveDown(n int) {
	if n < 1 {
		return
	}
	w.WriteString(fmt.Sprintf("\033[%dB", n))
}
func (w *Writer) MoveLeft(n int) {
	if n < 1 {
		return
	}
	w.WriteString(fmt.Sprintf("\033[%dD", n))
}
func (w *Writer) MoveRight(n int) {
	if n < 1 {
		return
	}
	w.WriteString(fmt.Sprintf("\033[%dC", n))
}
