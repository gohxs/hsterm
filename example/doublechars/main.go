package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gohxs/hsterm/ansi"
)

func main() {

	//Testing "象形字"

	exStdin()
	exAnsiReader()

}

func exStdin() {
	b := make([]byte, 100)
	// Check
	fmt.Print("stdin> ")
	_, err := os.Stdin.Read(b)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Readed:", string(b))

}

func exAnsiReader() {
	b := make([]byte, 100)
	fmt.Print("ansireader> ")
	ar := ansi.NewReader(os.Stdin)
	_, err := ar.Read(b)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Readed:", string(b))

}
