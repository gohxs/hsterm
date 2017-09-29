package main

import (
	"log"
	"runtime"

	"github.com/gohxs/termu"
)

func main() {
	log.Println("runtime.GOOS: ", runtime.GOOS)

	t := termu.New()
	t.SetPrompt("Win> ")

	s := termu.NewScanner(t)
	for s.Scan() {
		ln := s.Text()
		log.Println("Bye:", ln)
	}

}
