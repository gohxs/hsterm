package main

import (
	"bytes"
	"flag"
	"fmt"
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
	"github.com/gohxs/hsterm"
	"github.com/gohxs/hsterm/term"
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
	flag.BoolVar(&dbgFlag, "dbg", false, "Debug toggle")
	flag.BoolVar(&dbgTmux, "tmux", true, "Use tmux for debugging")
	flag.Parse()

	hsterm.DebugOutput = "out.txt"

	log.Println("Hello world")
	rl := hsterm.New()
	rl.AutoComplete = completeFunc
	rl.Display = display
	rl.SetPrompt("stdio@hsterm ~$ ")
	log.SetFlags(0)

	done := make(chan int, 1)

	// DEBUG UTILITY
	if dbgFlag {
		if dbgTmux {
			cmd := exec.Command("tmux", "split", "-h", "-p", "70", "-d", "tail -f out.txt")
			cmd.Stdin = os.Stdin
			cmd.Stderr = rl
			cmd.Stdout = rl
			cmd.Run()
		}
		// Capture signal
		insign := make(chan os.Signal)
		signal.Notify(insign, os.Interrupt)
		go func() { <-insign; close(done) }() // Signaler
		defer func() {                        // Done
			log.Println("Killing pane")
			if dbgTmux {
				c := exec.Command("tmux", "kill-pane", "-t", "2")
				c.Run()
			}
		}()
	}

	log.SetOutput(rl) // Change default logger to term

	//Progress bar helper
	progressBar := func() {
		count := 100000
		bar := pb.StartNew(count)
		bar.Output = rl
		for i := 0; i < count; i++ {
			bar.Increment()
			time.Sleep(20 * time.Microsecond)
		}
		bar.FinishPrint("The End!")
	}

	reSplitter := regexp.MustCompile("\\s+")
	// CHANS
	for {
		line, err := rl.ReadLine()
		if err != nil {
			close(done)
			break
		}

		cmds := reSplitter.Split(line, -1)
		fmt.Fprintln(rl, "line:", line, cmds) // Echoer

		switch cmds[0] {
		case "char":
			if len(cmds) < 2 {
				fmt.Fprintln(rl, "Wrong number of arguments")
				fmt.Fprintln(rl, "  Usage: char <c> [number]")
				continue
			}
			char := cmds[1]
			amount := term.GetScreenWidth() * 2
			if len(cmds) > 2 {
				amount, _ = strconv.Atoi(cmds[2])
			}
			go func() {
				for i := 0; i < amount; i++ {
					fmt.Fprint(rl, char)
					<-time.After(20 * time.Millisecond)
				}
			}()
		case "a":
			fmt.Fprint(rl, strings.Join(cmds, " "))
		case "line":
			if len(cmds) < 2 {
				fmt.Fprintln(rl, "Wrong number of arguments")
				fmt.Fprintln(rl, "  Usage: line <c> [number]")
				continue
			}
			ln := cmds[1]
			amount := term.GetScreenWidth() * 2
			if len(cmds) > 2 {
				amount, _ = strconv.Atoi(cmds[2])
			}
			go func() {
				for i := 0; i < amount; i++ {
					fmt.Fprintln(rl, ln)
					<-time.After(300 * time.Millisecond)
				}
			}()
		case "pb":
			if len(cmds) > 1 && cmds[1] == "background" {
				go progressBar()
				continue
			}
			progressBar()

		}
	}
	<-done

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
	if len(match) == 0 { // No match
		return "", 0, false
	}
	prefix := ""
indexFor:
	for i := 0; ; i++ {
		if i >= len(match[0]) {
			break indexFor
		}

		ch := match[0][i]
		for _, v := range match[1:] {
			if i > len(v) || ch != v[i] {
				break indexFor
			}
		}
		prefix += string(ch)
	}

	// Find smaller?
	log.Println()
	log.Println(match)

	if len(prefix) > 0 {
		if len(match) == 1 {
			prefix += " "
		}
		line = prefix
		pos = len(prefix)
		return prefix, pos, true
	}

	return line, pos, false
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
