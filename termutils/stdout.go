// +build darwin dragonfly freebsd linux,!appengine netbsd openbsd solaris

package termutils

// This is not a terminal, just a writer
import (
	"io"
	"os"
)

//NewStdoutWriter returns a prepared stdout
//  just passthrough
func NewStdoutWriter(rc io.Writer) (io.Writer, error) {
	f, ok := rc.(*os.File)
	if !ok {
		return nil, ErrNotTerminal
	}
	if !IsTerminal(int(f.Fd())) {
		return nil, ErrNotTerminal
	}
	return rc, nil
}
