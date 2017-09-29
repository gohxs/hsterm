package main

import (
	"fmt"
	"log"

	"github.com/gohxs/termu"
)

func main() {

	t := termu.New()
	t.SetPrompt("$ ")

	scanner := termu.NewScanner(t)
	//scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Println("Scanner error:", err)
	}
}
