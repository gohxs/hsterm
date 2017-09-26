// +build darwin dragonfly freebsd linux,!appengine netbsd openbsd solaris

package term

// This is not a terminal, just a writer
import (
	"io"
)

//NewStdoutWriter returns a prepared stdout
//  just passthrough
func NewStdoutWriter(rc io.Writer) io.Writer {
	return rc
}
