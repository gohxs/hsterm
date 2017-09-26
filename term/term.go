package term

import (
	"errors"
)

const (
	VariantUnix = iota
	VariantWindows
)

var (
	ErrNotTerminal    = errors.New("Not a terminal")
	ErrNotImplemented = errors.New("Not implemented")
)
