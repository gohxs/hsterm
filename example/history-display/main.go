package main

import (
	"bytes"
	"fmt"
	"log"
	"strings"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/quick"
	"github.com/alecthomas/chroma/styles"
	"github.com/gohxs/hsterm"
)

var (
	tab = 0
)

func main() {

	t := hsterm.New()
	t.SetPrompt("hello> ")
	t.Display = buildDisplay(t)
	t.AutoComplete = buildComplete(t)

	for {
		line, err := t.ReadLine()
		if err != nil {
			log.Println("err:", err)
			return
		}
		if strings.HasPrefix(line, "hist ") {
			fmt.Fprintln(t, strings.Join(t.History.List(), "\n"))
		}

		fmt.Fprintln(t, line)

	}
}

func histMatchList(t *hsterm.Term, in string) []string {
	ret := []string{}
	histList := t.History.List()
	for i := len(histList) - 1; i >= 0; i-- {
		v := histList[i]              // reverse
		if strings.HasPrefix(v, in) { // Print in in white, rest in black
			ret = append(ret, v)
		}
	}
	return ret
}

func buildDisplay(t *hsterm.Term) func(string) string {
	return func(in string) string {
		if len(in) == 0 { // pass right throuh its a 0
			return in
		}
		first, rest := "", ""
		n := utilSplitN(in, " ", &first, &rest)
		if n == 2 {
			rest = " " + rest
		}

		list := histMatchList(t, in)
		if len(list) == 0 {
			//sub display here
			return highlight(first + rest)
			//return "\033[01;31m" + first + "\033[0;36m" + rest + "\033[m"
		}
		m := list[tab%len(list)] // Select one from list
		return highlight(first+rest) + "\033[01;30m" + m[len(in):] + "\033[m"
		//return "\033[01;37m" + first + "\033[0;36m" + rest + "\033[01;30m" + m[len(in):] + "\033[m"
	}

}

func buildComplete(t *hsterm.Term) func(string, int, rune) (string, int, bool) {
	return func(line string, pos int, key rune) (newLine string, newPos int, ok bool) {
		if key != '\t' {
			tab = 0
			return
		}
		tab++
		if len(line) == 0 {
			return
		}

		// Prefix match
		list := histMatchList(t, line)
		if len(list) == 0 { // no match just return
			return
		}

		res := list[0]
		if len(list) > 1 { // one match only
			res = ""
			// Complete the common chars in list
		colFor:
			for i := 0; ; i++ {
				if i >= len(list[0]) {
					break
				}
				c := list[0][i]
				for _, v := range list[1:] {
					if i >= len(v) || c != v[i] {
						break colFor
					}
				}
				res += string(c)
			}
		}

		// Complete only until space if we are in a space move forward
		for pos < len(res) && res[pos] == ' ' { // CountSpace
			pos++
		}
		space := strings.Index(res[pos:], " ") // current position forward

		if space > 0 { // only positive space
			res = res[:pos+space]
		}
		if res != line { // reset tab if line changed
			tab = 0
		}
		// Go to next space only
		return res, len(res), true
	} // End complete
}

func utilSplitN(s string, sep string, targets ...*string) (n int) {
	N := len(targets)
	res := strings.SplitN(s, sep, N)
	n = len(res)
	for i, t := range targets {
		if i >= len(res) {
			return
		}
		*t = res[i]
	}
	return
}

func highlight(input string) string {
	buf := bytes.NewBuffer([]byte{})
	err := quick.Highlight(buf, input, "bash", "terminal16m", "monokaim")
	//err := quick.Highlight(buf, input, "bash", "terminal16m", "monokaim")
	if err != nil {
		log.Fatal(err)
	}
	return buf.String()
}

var _ = styles.Register(chroma.MustNewStyle("monokaim", chroma.StyleEntries{
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
