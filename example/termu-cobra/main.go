// Testing add a cobra command line parser to termu
package main

import (
	"log"
	"strings"

	"github.com/gohxs/termu"
	"github.com/spf13/cobra"
)

func main() {

	cmd := setupCobra()
	/*if err := cmd.Execute(); err != nil {
		cmd.Println("err:", err)
	}*/

	t := termu.New()
	t.AutoComplete = completer(cmd)
	cmd.SetOutput(t)

	cmdStack := []*cobra.Command{}
	cmdStack = append(cmdStack, cmd)
	t.SetPrompt("cobra> ")
	for {
		line, err := t.ReadLine()
		if err != nil {
			log.Fatal(err)
		}

		args := strings.Split(line, "\n")
		cmd.Println("Split:", args)

		cmd.SetArgs(args)
		err = cmd.Execute()
		if err != nil {
			cmd.Println(err)
		}

	}
}

func completer(cmd *cobra.Command) func(line string, pos int, key rune) (newLine string, newPos int, ok bool) {
	return func(line string, pos int, key rune) (newLine string, newPos int, ok bool) {
		if key != '\t' {
			return
		}
		cmd.Println("Listing")
		cmplList := []string{}
		list := cmd.Commands()
		for _, v := range list {
			cmplList = append(cmplList, v.Name())
		}
		cmd.Println("Complete:", cmplList)

		return
	}
}

func setupCobra() *cobra.Command {
	rootCmd := cobra.Command{}
	helloCmd := cobra.Command{
		Use: "hello",
		Run: func(cmd *cobra.Command, args []string) {
			log.Println("World")
		},
	}
	rootCmd.AddCommand(&helloCmd)

	return &rootCmd

}
