package hsterm

import "github.com/gohxs/hsterm/internal/term"

func init() {
	Stdin = term.NewRawReader()
	Stdout = term.New(Stdout)
}
