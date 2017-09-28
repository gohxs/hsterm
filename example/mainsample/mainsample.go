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
	t.AutoComplete = buildAutoComplete(t)
	t.Display = buildDisplay(t)
	t.Log = prettylog.New("", dbgFile) // Debug logger that sends to F

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

// Global tab
var tab = 0

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

func buildDisplay(t *termu.Term) func(string) string {
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

func buildAutoComplete(t *termu.Term) func(string, int, rune) (string, int, bool) {
	return func(line string, pos int, key rune) (newLine string, newPos int, ok bool) {
		if key != '\t' {
			tab = 0
			return
		}
		tab++

		buf := bytes.NewBuffer(nil)

		list := histMatchList(t, line)
		fmt.Fprint(buf, fmt.Sprintf("\033[%dB", t.PromptLineCount()+1))
		for i, v := range list {
			//utilSplitN(v, " ", &first) // Parse scan??
			if i == tab%len(list) {
				fmt.Fprintf(buf, "\033[7m")
			}
			fmt.Fprintf(buf, "%s\033[m\t", v)
		}
		fmt.Fprintln(buf)

		t.Write(buf.Bytes())
		//if len(line) == 0 { // Maybe we can show a list splited by first char
		//	return
		//}

		// Prefix match
		//list := histMatchList(t, line)
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
			res = res[:pos+space] + " "
		}
		if res != line { // reset tab if line changed
			tab = 0
		}
		// Go to next space only
		return res, len(res), true
	} // End complete
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
