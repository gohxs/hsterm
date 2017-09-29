package termutils_test

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gohxs/termu/termutils"
)

type coords struct {
	x, y   int    // Expected coords
	escape string // Escape
	label  string // Label
	redraw bool   // Redraw the guidelines
}

var (
	height      = 30
	colorCoords = []coords{
		{55, 15, "\033[15;55H\033[46m\033[2K", "To 15;55 and Clear Line with color", true},
		{55, 17, "\033[2B\033[44m\033[2J", "2Down and clear fullscreen with color", false},
	}
	movementCoords = []coords{ // Movement
		{10, 10, "\033[10;10H", "Move to line 10, col 10", true},
		{11, 10, "\033[C", "1 right", false},
		{11, 12, "\033[2B", "2 Down", false},
		{9, 12, "\033[2D", "2 Left", false},
		{9, 10, "\033[2A", "2 Up", false},
		{10, 2, "\033[2;10H", "Move to line 2, col 10", false},
		{10, 6, "\033[4B", "Move down 4 times", false},
		{10, 7, "\033[0B", "Move down 1 times", false},
		{1, 5, "\033[5H", "Move to fifth line", false},
	}
	lineCleaningCoords = []coords{ // Line cleaning
		{30, 5, "\033[5;30H\033[K", "To line 5, col 30, Clear Line to EOL", true},
		{30, 7, "\033[2B\033[1K", "Move 2 down, and clear to begin", false},
		{30, 10, "\033[3B\033[2K", "Move 3 down, and clear line", false},
		{30, 15, "\033[5B\033[01;35mSome color here\033[47;30mHI background color\033[0m", "5 Down and print text", false},
		{4, 15, "\033[15;4H\033[K", "Line 15 and clear to right", false},
	}
	terminalCleaningCoords = []coords{ // Terminal cleaning
		{45, 15, "\033[15;45H\033[J", "line 15, col 45, Clear down", true},

		{45, 17, "\033[17;45H\033[1J", "line 17, col 45, Clear Up", true},

		{45, 20, "\033[20;45H\033[2J", "line 20, col 45, Clear ALL", true},
	}
)

func TestNonTerminal(t *testing.T) {

	f, err := ioutil.TempFile(os.TempDir(), "term-test")
	if err != nil {
		t.Fatal(err)
	}

	_, err = termutils.NewStdoutWriter(f)
	if err == nil {
		t.Fatal("Output should fail, because its not a terminal")
	}

}

func TestSaveRestore(t *testing.T) {
	tw, err := termutils.NewStdoutWriter(os.Stdout)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Fprintln(tw, "\033[2J\033[H")
	fmt.Fprint(tw, "Saving cursor here >\033[s <\n")
	fmt.Fprint(tw, "Do more things\n")
	<-time.After(3 * time.Second)
	fmt.Fprint(tw, "Restoring cursor there\n")
	fmt.Fprint(tw, "\033[u")
	<-time.After(3 * time.Second)

}
func TestCursor(t *testing.T) {
	tw, err := termutils.NewStdoutWriter(os.Stdout)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Fprintln(tw, "\033[2J\033[H")
	fmt.Fprintln(tw, "Hiding cursor for 3sec\033[?25l")
	<-time.After(3 * time.Second)
	fmt.Fprintln(tw, "do things, show in 3 sec")
	<-time.After(3 * time.Second)
	fmt.Fprintln(tw, "\033[?25h")
	<-time.After(3 * time.Second)
}

func TestColors(t *testing.T) {
	tw, err := termutils.NewStdoutWriter(os.Stdout)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Fprintln(tw, "\033[2J\033[HColor test:")
	fmt.Fprintln(tw, "\033[30;42;46;31mMulti param\033[0m")
	fmt.Fprintln(tw, "\033[30mBLACK\033[01;30mBLACK\033[40mBLACK\033[0m rest")
	fmt.Fprintln(tw, "\033[31mRED\033[01;31mRED\033[41mRED\033[0m rest")
	fmt.Fprintln(tw, "\033[32mGREEN\033[01;32mGREEN\033[42mGREEN\033[0m rest")
	fmt.Fprintln(tw, "\033[33mYELLOW\033[01;33mYELLOW\033[43mYELLOW\033[0m rest")
	fmt.Fprintln(tw, "\033[34mBLUE\033[01;34mBLUE\033[44mBLUE\033[0m rest")
	fmt.Fprintln(tw, "\033[35mMAGENTA\033[01;35mMAGENTA\033[45mMAGENTA\033[0m rest")
	fmt.Fprintln(tw, "\033[36mCYAN\033[01;36mCYAN\033[46mMAGENTA\033[0m rest")
	fmt.Fprintln(tw, "\033[37mWHITE\033[01;37mWHITE\033[47mWHITE\033[0m rest")
	fmt.Fprintln(tw, "\033[31mcolor\033[7mreverse\033[34mcolorchange\033[m rest")

	fmt.Fprintln(tw, "\033[01mcolor\033[m\033[30m\033[47m Separated\033[m")

	// 256 colors
	for i := 0; i < 255; i++ {
		if i%30 == 0 {
			fmt.Fprintf(tw, "\n")
		}
		fmt.Fprintf(tw, "\033[38;5;%dm#", i)
	}
	fmt.Fprint(tw, "\033[0m\n")

}

func TestFullColorClear(t *testing.T) {
	runTest(t, colorCoords)
}

////////////////////
// TEST FUNC
func TestMovement(t *testing.T) {
	runTest(t, movementCoords)
}
func TestLineClean(t *testing.T) {
	runTest(t, lineCleaningCoords)
}
func TestTerminalClean(t *testing.T) {
	runTest(t, terminalCleaningCoords)
}

func runTest(t *testing.T, curCoords []coords) {
	if testing.Short() {
		t.Skipf("Test is not short")
		return
	}
	tw, err := termutils.NewStdoutWriter(os.Stdout)
	if err != nil {
		t.Fatal(err)
	}

	for curIndex, c := range curCoords {
		if c.redraw {
			<-time.After(3 * time.Second)
			// Wait
			fmt.Fprintf(tw, screen(curIndex, curCoords))
			//hPrint(tw, curIndex+1, screen(curIndex, curCoords), "Redraw")
		}
		hPrint(tw, curIndex, curCoords)
		//hPrint(tw, curIndex+1, c.escape, c.label)
		fmt.Fprintf(tw, "\033[0m")
		<-time.After(1000 * time.Millisecond)
	}
	<-time.After(1 * time.Second)
}

func hPrintOne(tw io.Writer, index int, test string, label string) {

	var status string
	if len(test) > 10 {
		status = fmt.Sprintf("\033[31H\033[32m%d - %s\033[0m\033[K", index, label) // Manual clear
	} else {
		status = fmt.Sprintf("\033[31H\033[32m%d - %#v\033[35m : %s\033[0m\033[K", index, test, label) // Manual clear
	}

	for tickLeft := 3; tickLeft > 0; tickLeft-- {
		fmt.Fprintf(tw, "\033[s%s (%d)\033[u", status, tickLeft)
		<-time.After(time.Second)
	}
	fmt.Fprintf(tw, "\033[s%s (%d)\033[u", status, 0)

	fmt.Fprintf(tw, "%s\033[s", test)
	<-time.After(300 * time.Millisecond)
}

func hPrint(tw io.Writer, index int, testCoords []coords) {

	//curTest := testCoords[index]
	var status string
	status = fmt.Sprintf("\033[32H\033[m")
	for i, t := range testCoords { // Print all coords
		test := t.escape
		label := t.label
		if i == index {
			status += fmt.Sprintf("\033[44m\033[1m")
		}
		if len(test) > 20 { // Do not print if big
			status += fmt.Sprintf("\033[32m%d - \033[35m%s\033[0m\033[K\n", i+1, label) // Manual clear
		} else {
			status += fmt.Sprintf("\033[32m%d - %#v\033[35m : %s\033[0m\033[K\n", i+1, test, label) // Manual clear
		}
	}
	test := testCoords[index].escape // ??

	tickLeft := 3
	for ; tickLeft > 0; tickLeft-- {
		fmt.Fprintf(tw, "\033[31H\033[m\033[2KTesting in \033[01m%d\033[0m %s\033[u", tickLeft, status)
		<-time.After(time.Second)
	}
	fmt.Fprintf(tw, "\033[31H\033[m\033[2KTesting in \033[01m%d\033[0m %s\033[u", tickLeft, status)

	fmt.Fprintf(tw, "\033[m%s\033[s", test)
}
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
