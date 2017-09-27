package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/gohxs/termu"
)

func main() {

	data, _ := ioutil.ReadFile("data/ansi.vt")
	br := bytes.NewReader(data)

	t := termu.New()

	scanner := bufio.NewScanner(br)
	lineCount := 0
	for scanner.Scan() {
		t.Write([]byte(scanner.Text() + "\n"))
		lineCount++
		<-time.After(50 * time.Millisecond)
		if lineCount > 25 {
			lineCount = 0
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
	/**/
}
