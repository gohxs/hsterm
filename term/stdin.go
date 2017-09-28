// +build darwin dragonfly freebsd linux,!appengine netbsd openbsd solaris

package term

// This is not a terminal, just a reader
import (
	"io"
)

// StdinReader
// Should be a STDIN
func NewStdinReader(rc io.ReadCloser) io.ReadCloser {
	return rc
}
