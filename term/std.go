package term

import (
	"io"
	"os"
)

var (
	Stdin  io.ReadCloser  = os.Stdin
	Stdout io.WriteCloser = os.Stdout
)
