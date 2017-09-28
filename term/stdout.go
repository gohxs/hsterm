// +build darwin dragonfly freebsd linux,!appengine netbsd openbsd solaris

package term

// This is not a terminal, just a writer
import (
	"io"
	"os"

	"github.com/gohxs/termu/term/termutils"
)

//NewStdoutWriter returns a prepared stdout
//  just passthrough
func NewStdoutWriter(rc io.Writer) (io.Writer, error) {
	f, ok := rc.(*os.File)
	if !ok {
		return nil, ErrNotTerminal
	}
	if !termutils.IsTerminal(int(f.Fd())) {
		return nil, ErrNotTerminal
	}
	return rc, nil
}
