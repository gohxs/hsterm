package termu

import (
	"fmt"
	"io"
	"runtime"
	"strings"
	"time"

	_ "github.com/gohxs/prettylog"
)

func min(n, min int) int {
	if n < min {
		return min
	}
	return n
}

func max(n, max int) int {
	if n > max {
		return max
	}
	return n
}

var (
	SlowRWDelay = 100
)

type SlowReader struct {
	io.Reader
}

func (sr SlowReader) Read(b []byte) (int, error) {
	<-time.After(time.Duration(SlowRWDelay) * time.Millisecond)
	return sr.Reader.Read(b)
}

type SlowWriter struct {
	io.Writer
}

func (sw SlowWriter) Write(b []byte) (int, error) {
	<-time.After(time.Duration(SlowRWDelay) * time.Millisecond)
	return sw.Writer.Write(b)
}

func caller(rel ...int) string {
	def := 2
	if len(rel) > 0 {
		def -= rel[0]
	}
	ptr, file, line, _ := runtime.Caller(def)

	tname := runtime.FuncForPC(ptr).Name()
	method := tname[strings.LastIndex(tname, ".")+1:]
	fname := file[strings.LastIndex(file, "/")+1:]

	return fmt.Sprintf("%s:%d/%s", fname, line, method)

}
