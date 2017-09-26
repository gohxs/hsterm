package main

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gohxs/hsterm"
)

type coords struct {
	x, y   int    // Expected coords
	escape string // Escape
	label  string // Label
	redraw bool   // Redraw the guidelines
}

var (
	height = 30
	tests  = [][]coords{
		[]coords{ // Colors
			{20, 20, "\033[20;20H\033[31mThis is a RED string", "Print RED", true},
		},
		[]coords{ // Movement
			{10, 10, "\033[10;10H", "Move to line 10, col 10", true},
			{11, 10, "\033[C", "1 right", false},
			{11, 12, "\033[2B", "2 Down", false},
			{9, 12, "\033[2D", "2 Left", false},
			{9, 10, "\033[2A", "2 Up", false},
			{10, 2, "\033[2;10H", "Move to line 2, col 10", false},
			{10, 6, "\033[4B", "Move down 4 times", false},
			{10, 7, "\033[0B", "Move down 1 times", false},
			{1, 5, "\033[5H", "Move to fifth line", false},
		},
		[]coords{ // Line cleaning
			{30, 5, "\033[5;30H\033[K", "To line 5, col 30, Clear Line to EOL", true},
			{30, 7, "\033[2B\033[1K", "Move 2 down, and clear to begin", false},
			{30, 10, "\033[3B\033[2K", "Move 3 down, and clear line", false},
		},
		[]coords{ // Terminal cleaning
			{45, 15, "\033[15;45H\033[J", "line 15, col 45, Clearn down", true},

			{45, 17, "\033[17;45H\033[1J", "line 17, col 45, Clear Up", true},

			{45, 20, "\033[20;45H\033[2J", "line 20, col 45, Clear ALL", true},
		},
	}
)

func screen(curIndex int, testCoords []coords) string {

	buf := bytes.NewBuffer(nil)
	buf.WriteString("\033[2J\033[H") // Clear
	for y := 0; y < height; y++ {    // Line by line
		line := fmt.Sprintf("%2d %s %2d\n", y+1, strings.Repeat("-", 100), y+1)

		for x := 0; x < len(line); x++ {
			found := 0
			for i, c := range testCoords { // Find coords
				if c.x == (x+1) && c.y == (y+1) {
					found = i + 1
					break
				}
			}
			if found > 0 {
				buf.WriteString(fmt.Sprintf("\033[01;3%dm%d\033[m", (found/10 + 2), found%10))
			} else {
				buf.WriteString(string(line[x]))
			}
		}
	}
	return buf.String()
}

func main() {

	t := hsterm.New()

	log.Println("Height:", height)

	hPrint := func(test string, label string) {
		for tickLeft := 5; tickLeft > 0; tickLeft-- {
			fmt.Fprintf(t, "\033[s")
			if len(test) > 10 {
				fmt.Fprintf(t, "\033[31H\033[32m%s\033[0m (%d)\033[K", label, tickLeft) // Manual clear
			} else {
				fmt.Fprintf(t, "\033[31H\033[32m%#v\033[35m : %s\033[0m (%d)\033[K", test, label, tickLeft) // Manual clear
			}
			fmt.Fprintf(t, "\033[u")
			<-time.After(500 * time.Millisecond)
		}
		fmt.Fprintf(t, "%s\033[s\033[01;31m\033[0m\033[u", test)
		<-time.After(1 * time.Second)
	}

	fmt.Fprintln(t, "Color test:")
	fmt.Fprintln(t, "\033[30;42;46;31mMulti param\033[0m")
	fmt.Fprintln(t, "\033[30mBLACK\033[01;30mBLACK\033[40mBLACK\033[0m rest")
	fmt.Fprintln(t, "\033[31mRED\033[01;31mRED\033[41mRED\033[0m rest")
	fmt.Fprintln(t, "\033[32mGREEN\033[01;32mGREEN\033[42mGREEN\033[0m rest")
	fmt.Fprintln(t, "\033[33mYELLOW\033[01;33mYELLOW\033[43mYELLOW\033[0m rest")
	fmt.Fprintln(t, "\033[34mBLUE\033[01;34mBLUE\033[44mBLUE\033[0m rest")
	fmt.Fprintln(t, "\033[35mMAGENTA\033[01;35mMAGENTA\033[45mMAGENTA\033[0m rest")
	fmt.Fprintln(t, "\033[36mCYAN\033[01;36mCYAN\033[46mMAGENTA\033[0m rest")
	fmt.Fprintln(t, "\033[37mWHITE\033[01;37mWHITE\033[47mWHITE\033[0m rest")
	<-time.After(5 * time.Second)
	// 256 colors
	for i := 0; i < 255; i++ {
		if i%30 == 0 {
			fmt.Fprintf(t, "\n")
		}
		fmt.Fprintf(t, "\033[38;5;%dm#", i)
	}
	fmt.Fprint(t, "\033[0m\n")
	<-time.After(5 * time.Second)

	for _, curCoords := range tests {
		for curIndex, c := range curCoords {
			if c.redraw {
				hPrint(screen(curIndex, curCoords), "Redraw")
			}
			hPrint(c.escape, c.label)
			fmt.Fprintf(t, "\033[0m")
		}

	}
	// Line 3
	//t.Flush()

}
