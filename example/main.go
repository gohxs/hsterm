package main

import (
	"bytes"
	"fmt"
	"io"
	glog "log"
	"strings"

	"github.com/alecthomas/chroma/quick"
	"github.com/gohxs/hsterm"
	"github.com/gohxs/prettylog"
)

type compl struct {
	token string
	child []*compl
}

var (
	log      *glog.Logger
	complete = []*compl{
		{token: "test1"},
		{token: "test123"},
	}
	out io.Writer
)

func main() {
	term := hsterm.New()
	term.AutoComplete = completeFunc
	term.Display = display

	term.Prompt("t> ")
	out = term

	defer term.Close()
	log = prettylog.New("term", term)

	//go func() {
	//for {
	//fmt.Fprintln(term, "Hello!")
	//<-time.After(4 * time.Second)
	//}
	//}()

	for {
		line, err := term.Readline()
		if err != nil {
			break
		}
		log.Println(line)
	}
}

func completeFunc(line string, pos int, key rune) (newLine string, newPos int, ok bool) {
	if key != '\t' {
		return
	}
	var match []string
	for _, v := range complete {
		if strings.HasPrefix(v.token, line) {
			match = append(match, v.token)
		}
	}
	// Find smaller?
	fmt.Fprintln(out, "M:", match)

	log.Println("Postion is:", pos)
	return line, pos, true
}
func display(input string) string {
	buf := bytes.NewBuffer([]byte{})
	err := quick.Highlight(buf, input, "postgres", "terminal16m", "monokaim")
	//err := quick.Highlight(buf, input, "bash", "terminal16m", "monokaim")
	if err != nil {
		log.Fatal(err)
	}
	return buf.String()
}
