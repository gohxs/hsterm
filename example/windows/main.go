package main

import (
	"log"
	"os"
	"runtime"

	"github.com/gohxs/termu/term/termutils"
)

func main() {
	log.Println("runtime.GOOS: ", runtime.GOOS)

	state, err := termutils.GetState(int(os.Stdin.Fd()))

	state.Mode |= 0x2000

}
