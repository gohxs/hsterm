package ansi

import (
	"fmt"
	"io"
	"os"
)

type DebugWriter struct {
	io.Writer
	outfile string
}

func NewDebugWriter(wr io.Writer, outfile string) io.Writer {
	return DebugWriter{wr, outfile}
}

func (dw DebugWriter) Write(b []byte) (int, error) {
	f, _ := os.OpenFile(dw.outfile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.FileMode(0644))
	defer f.Close()
	fmt.Fprintf(f, "%#v\n", string(b))
	return dw.Writer.Write(b)
}
