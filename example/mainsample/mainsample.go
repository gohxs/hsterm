package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/quick"
	"github.com/alecthomas/chroma/styles"
	"github.com/cheggaaa/pb"
	"github.com/gohxs/prettylog"
	"github.com/gohxs/termu"
)

type compl struct {
	token string
	child []*compl
}

var (
	complete = []*compl{
		{token: "test1"},
		{token: "test123"},
	}

	dbgFlag bool
	dbgTmux bool
)

func main() {
	// Debugs:
	//return
	flag.BoolVar(&dbgFlag, "dbg", false, "Debug toggle")
	flag.BoolVar(&dbgTmux, "tmux", true, "Use tmux for debugging")
	flag.Parse()

	dbgFile, err := os.OpenFile("dbg.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.FileMode(0644))
	if err != nil {
		panic(err)
	}

	done := make(chan int, 1)

	// TERM UTILS SETUP
	t := termu.New()
	t.SetPrompt("stdio@termu ~$ ")
	t.Log = prettylog.New("", dbgFile) // Debug logger that sends to F

	ce := ComplEngine{term: t}
	t.AutoComplete = ce.AutoComplete
	t.Display = ce.Display

	t.History.Append("char - 1000")
	t.History.Append("pb &")
	t.History.Append("pb")

	// DEBUG UTILITY
	if dbgFlag {
		if dbgTmux {
			cmd := exec.Command("tmux", "split", "-h", "-p", "70", "-d", "tail -f dbg.log")
			cmd.Stdin = os.Stdin
			cmd.Stderr = os.Stderr
			cmd.Stdout = os.Stdout
			err := cmd.Run()
			log.Println("Tmux err", err)

			defer func() { // Done
				c := exec.Command("tmux", "kill-pane", "-t", "2")
				c.Run()
			}()
		}
		// Capture signal

	}
	insign := make(chan os.Signal)
	signal.Notify(insign, os.Interrupt)
	go func() { <-insign; close(done) }() // Signaler

	reSplitter := regexp.MustCompile("\\s+")
	// CHANS
	for {
		line, err := t.ReadLine()
		if err != nil {
			close(done)
			break
		}
		cmds := reSplitter.Split(line, -1)

		switch cmds[0] {
		case "echo":
			fmt.Fprint(t, strings.Join(cmds[1:], " "))
		case "char", "line":
			if len(cmds) < 2 {
				fmt.Fprintln(t, "Wrong number of arguments")
				fmt.Fprintf(t, "  Usage: %s <c> [number]\n", cmds[0])
				continue
			}
			asyncRepeat(t, cmds)
		case "pb":
			if len(cmds) > 1 && cmds[1] == "&" {
				go progressBar(t)
				continue
			}
			progressBar(t)

		}
	}
	<-done
}

// Something to print
func asyncRepeat(t *termu.Term, cmds []string) {
	delay := 20
	if cmds[0] == "line" {
		cmds[1] += "\n"
		delay = 800
	}
	char := cmds[1]
	amount, _ := t.GetSize()
	amount *= 2
	if len(cmds) > 2 {
		amount, _ = strconv.Atoi(cmds[2])
	}
	go func() {
		for i := 0; i < amount; i++ {
			fmt.Fprint(t, char)
			<-time.After(time.Duration(delay) * time.Millisecond)
		}
	}()
}

type ComplEngine struct {
	term    *termu.Term
	tab     int
	mode    int
	suggest []string // suggestions, maybe not needed
}

func (c *ComplEngine) Display(in string) string {
	if len(in) == 0 { // pass right throuh its a 0
		return in
	}
	first, rest := "", ""
	n := utilSplit(in, " ", &first, &rest)
	if n == 2 {
		rest = " " + rest
	}
	if len(c.suggest) == 0 {
		return highlight(first + rest)
	}
	m := c.suggest[c.tab%len(c.suggest)] // Select one from list
	res := highlight(first+rest) + "\033[01;30m" + m[len(in):] + "\033[m"

	if c.mode != 1 {
		return res
	}
	res += "\n"
	for i, v := range c.suggest {
		var treated string
		fmt.Sscanf(v[len(in):], "%s", &treated)
		if i == c.tab%len(c.suggest) {
			res += "\033[7m" + treated + "\t\033[m"
			continue
		}
		res += treated + "\t"
	}
	return res
}

// Any typing should show this right?
// Hijacking function
func (c *ComplEngine) AutoComplete(line string, pos int, key rune) (newLine string, newPos int, ok bool) {
	if key == '\n' && c.mode == 1 { // Wont happen?
		c.mode = 0
		if len(c.suggest) == 0 {
			return
		}
		sugLine := c.suggest[c.tab%len(c.suggest)]
		if line == sugLine {
			return
		}
		newLine = sugLine
		newPos = len(newLine)
		ok = true
		return
	}

	if key != '\t' {
		c.suggest = histMatchList(c.term, line+string(key)) // On type
		c.tab = 0
		return
	}
	c.mode = 1
	c.tab++

	c.suggest = histMatchList(c.term, line) // On type
	if len(c.suggest) == 0 {                // no match just return
		return
	}

	res := c.suggest[0]
	if len(c.suggest) > 1 { // one match only
		res = ""
		// Complete the common chars in list
	colFor:
		for i := 0; ; i++ {
			if i >= len(c.suggest[0]) {
				break
			}
			cur := c.suggest[0][i]
			for _, v := range c.suggest[1:] {
				if i >= len(v) || cur != v[i] {
					break colFor
				}
			}
			res += string(cur)
		}
	}

	// Complete only until space if we are in a space move forward
	for pos < len(res) && res[pos] == ' ' { // CountSpace
		pos++
	}
	space := strings.Index(res[pos:], " ") // current position forward

	if space > 0 { // only positive space
		res = res[:pos+space] + " "
	}
	if res != line { // reset tab if line changed
		c.tab = 0
	}
	// Go to next space only
	return res, len(res), true
	// End complete
}

func histMatchList(t *termu.Term, in string) []string {
	if len(in) == 0 {
		return t.History.List()
	}
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

func highlight(input string) string {
	buf := bytes.NewBuffer([]byte{})
	err := quick.Highlight(buf, input, "sql", "terminal", "monokaim")
	//err := quick.Highlight(buf, input, "bash", "terminal16m", "monokaim")
	if err != nil {
		log.Fatal(err)
	}
	return buf.String()
}
func progressBar(w io.Writer) { //Progress bar helper
	count := 100000
	bar := pb.StartNew(count)
	bar.Output = w
	for i := 0; i < count; i++ {
		bar.Increment()
		time.Sleep(20 * time.Microsecond)
	}
	bar.FinishPrint("The End!")
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

// Utility
func utilSplit(s string, sep string, targets ...*string) (n int) {
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
