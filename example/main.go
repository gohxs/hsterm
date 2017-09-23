package main

import (
	"bytes"
	"fmt"
	"io"
	glog "log"
	"strings"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/quick"
	"github.com/alecthomas/chroma/styles"
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
	err := quick.Highlight(buf, input, "postgres", "terminal", "monokaim")
	//err := quick.Highlight(buf, input, "bash", "terminal16m", "monokaim")
	if err != nil {
		log.Fatal(err)
	}
	return buf.String()
}

var Monokai = styles.Register(chroma.MustNewStyle("monokaim", chroma.StyleEntries{
	chroma.Text:                "#f8f8f2",
	chroma.Error:               "#960050 bg:#1e0010",
	chroma.Comment:             "#75715e",
	chroma.Keyword:             "#66d9ef",
	chroma.KeywordNamespace:    "#f92672",
	chroma.Operator:            "#f92672",
	chroma.Punctuation:         "#f8f8f2",
	chroma.Name:                "#f8f8f2",
	chroma.NameBuiltin:         "#f844d0",
	chroma.NameAttribute:       "#a6e22e",
	chroma.NameClass:           "#a6e22e",
	chroma.NameConstant:        "#66d9ef",
	chroma.NameDecorator:       "#a6e22e",
	chroma.NameException:       "#a6e22e",
	chroma.NameFunction:        "#a6e22e",
	chroma.NameOther:           "#a6e22e",
	chroma.NameTag:             "#f92672",
	chroma.LiteralNumber:       "#ae81ff",
	chroma.Literal:             "#ae81ff",
	chroma.LiteralDate:         "#e6db74",
	chroma.LiteralString:       "#e6db74",
	chroma.LiteralStringEscape: "#ae81ff",
	chroma.GenericDeleted:      "#f92672",
	chroma.GenericEmph:         "italic",
	chroma.GenericInserted:     "#a6e22e",
	chroma.GenericStrong:       "bold",
	chroma.GenericSubheading:   "#75715e",
}))
