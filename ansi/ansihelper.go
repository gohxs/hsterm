// Wrong should be different, this is a helper

package ansi

import (
	"bytes"
	"fmt"
	"io"
)

// Maybe const this

//Helper struct that holds methods to easy write ansi escape codes
type Helper struct {
	buf *bytes.Buffer
	wr  io.Writer

	io.Writer
}

//NewHelperBuffered Buffered it will write to an internal buffer
// and will be sent to the writer on a Flush
// content is clear after a flush
func NewHelperBuffered(wr io.Writer) *Helper {
	b := bytes.NewBuffer(nil)
	return &Helper{buf: b, wr: wr, Writer: b}
}

//NewHelperDirect writes directly to writer (i.e. not buffered)
func NewHelperDirect(wr io.Writer) *Helper {
	return &Helper{Writer: wr}
}

func (w *Helper) Flush() error {
	_, err := w.wr.Write(w.buf.Bytes())
	w.buf.Reset()
	return err
}
func (w *Helper) String() string {
	return w.buf.String()
}

//WriteString write directly
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
	fmt.Fprintf(w, "\033[%dA", n)
}
func (w *Helper) MoveDown(n int) {
	if n < 1 {
		return
	}
	fmt.Fprintf(w, "\033[%dB", n)
}
func (w *Helper) MoveLeft(n int) {
	if n < 1 {
		return
	}
	fmt.Fprintf(w, "\033[%dD", n)
}
func (w *Helper) MoveRight(n int) {
	if n < 1 {
		return
	}
	fmt.Fprintf(w, "\033[%dC", n)
}

func (w *Helper) ClearLineToEnd() {
	w.WriteString("\033[K")
}
func (w *Helper) ClearLineFromStart() {
	w.WriteString("\033[1K")
}
func (w *Helper) ClearLine() {
	w.WriteString("\033[2K")
}

func (w *Helper) ClearScreenToEnd() {
	w.WriteString("\033[J")
}
func (w *Helper) ClearScreenFromStart() {
	w.WriteString("\033[1J")
}
func (w *Helper) ClearScreen() {
	w.WriteString("\033[2J")
}
