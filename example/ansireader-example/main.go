package main

import (
	"log"
	"os"

	"github.com/gohxs/hsterm/internal/ansireader"
)

func main() {

	rd := ansireader.New(os.Stdin)

	for {
		v, _ := rd.Read()
		log.Printf("Value: %#v", v)
	}

}
