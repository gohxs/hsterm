package main

import (
	"fmt"
	"io"

	"github.com/gohxs/hsterm/ansireader"
)

func main() {

	done := make(chan struct{})

	reader, writer := io.Pipe()
	go func() {
		data := make([]byte, 64)
		areader := ansireader.New(reader)
		for { // Do whatever parse the things whatever
			n, err := areader.Read(data) // Just echo
			fmt.Print(string(data[:n]))  // Echoer

			if err != nil {
				close(done)
				return
			}
		}
	}()

	//rd := ansireader.New(os.Stdin)
	//
	toWrite := "\033[01;43mThis should be a thing\033[m"
	fmt.Println(toWrite)

	fmt.Fprintln(writer, toWrite)
	writer.Close()

	<-done

}
