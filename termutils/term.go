package termutils

import (
	"errors"
)

var (
	ErrNotTerminal    = errors.New("Not a terminal")
	ErrNotImplemented = errors.New("Not implemented")
)
