package main

import (
	"bytes"
	"log"

	"github.com/alecthomas/chroma/quick"
	"github.com/gohxs/readline"
)

// Highlighter test
func display(input string) string {
	buf := bytes.NewBuffer([]byte{})
	err := quick.Highlight(buf, input, "postgres", "terminal16m", "monokai")
	if err != nil {
		log.Fatal(err)
	}
	return buf.String()
}
func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	term, err := readline.NewEx(&readline.Config{
		Prompt: "SQLIrl> ",
		Output: display,
	})
	if err != nil {
		log.Fatal(err)
	}
	for {
		line, err := term.Readline()
		if err != nil {
			log.Fatal(err)
		}
		log.Println("E:", line)
	}

}
