package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"golang.org/x/crypto/ssh/terminal"
)

type ioTerm struct {
	io.Reader
	io.Writer
}

func main() {

	tr := terminal.NewTerminal(ioTerm{os.Stdin, os.Stdout}, "term> ")
	for {
		line, err := tr.ReadLine()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Fprintln(tr, "Echo:", line)

	}
}
