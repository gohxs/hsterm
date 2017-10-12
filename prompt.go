package termu

// Prompt handles the prompt area of the readline,
// Including any extra things
import (
	"runtime"
	"strings"

	"github.com/gohxs/termu/ansi"
)

type Prompt struct {
	//Previous string so we can clean?
	*InputBuffer
	term      *Term // Term that prompt belongs too
	prompt    string
	extraLine int
}

func newPrompt(t *Term) *Prompt {
	return &Prompt{&InputBuffer{}, t, "", 0}
}

func (p *Prompt) SetPrompt(prompt string) {
	p.prompt = prompt
}

func (p *Prompt) String(a, b error) {}

//TODO: This should be private
//PromptString Returns the prompt string proper escaped
func (p *Prompt) DisplayString() string {
	// Transform output
	var dispBuf string
	// Idea cache this, every time we print the thing, we transform
	// if transformation is different than cache we redraw, else we put the difference
	if p.term.Display != nil {
		dispBuf = p.term.Display(p.InputBuffer.String())
	} else {
		dispBuf = p.InputBuffer.String()
	}
	// To properly count the lines
	// Split input buffer? and count each line
	//
	termWidth := min(p.term.width, 1) // Grab the terminal width or minimum 1
	rawDispBuf := ansi.Strip(dispBuf) // Strip any escape (color etc) // Carefull with the movement things

	lines := strings.Split(rawDispBuf, "\n")
	// count line length
	lineCount := (len(p.prompt)+len(lines[0]))/termWidth + 1 // this is the first line
	if len(lines) > 1 {
		for _, v := range lines[1:] { // next lines if any
			// Here we do the same to check the wrap line
			lineCount += len(v)/termWidth + 1
		}
	}
	// we also should check cursor line position here
	dispLen := len(rawDispBuf)
	fullPromptLen := dispLen + len(p.prompt)
	//lineCount := fullPromptLen/termWidth + 1

	// Count lines -- in rawDispBuf

	ab := ansi.NewHelperBuffered(nil)
	count := lineCount - 1 + p.extraLine // We need to check with the input

	ab.RestoreCursor() // Or prompt clear
	//For redraw
	if count > 0 { // Move back input and thing buffer?
		// Scroll hack
		// with restore cursor, if it is at last line,
		// it will scroll up any remaining text repeating the top line of the prompt
		// We line feed the number of lines prompt will have
		// but since the cursor will go to 0 position, we have to:
		// restore again,
		// As the cursor is at bottom line, so we moveDown (will not move if in bottom)
		// and move up back to desired position, saving the cursor for the next Write
		// if the cursor is not on the last line it will perform movements and
		// will end up same position
		ab.WriteString(strings.Repeat("\n", count)) // Form feed if necessary
		ab.RestoreCursor()                          // Move back
		ab.MoveDown(count)                          //
		ab.MoveUp(count)
		ab.SaveCursor() // Save cursor again
	} /**/

	// Transform output
	ab.WriteString(strings.Repeat("\n", p.extraLine))
	ab.WriteString(strings.Repeat("\033[K", lineCount)) // N lines

	// Here we are printing display
	ab.WriteString("\r" + p.prompt)
	ab.WriteString(dispBuf)

	// If not windows
	if runtime.GOOS != "windows" {
		ab.WriteString(" \b")
	}

	//XXX: if we add a special printer that prints bellow
	//screen it will becleared, so we need to only clear the LINE2
	//ab.WriteString("\033[K") // Clean Rest of the line
	ab.WriteString("\033[J") // Maybe rest of the line
	// Position cursor // Reposition cursor here??
	// No unicode for now
	cursorLen := fullPromptLen - (dispLen - p.InputBuffer.Cursor())
	clineCount := count - p.extraLine // Remove extra line

	desiredLine := (cursorLen / termWidth) // get Line position starting from prompt
	desiredCol := cursorLen % termWidth    // get column position starting from prompt

	p.term.Log.Println("Backing up:", clineCount, desiredLine)
	ab.MoveUp(clineCount) // Origin
	ab.WriteString("\r")  // go back anyway
	ab.MoveDown(desiredLine)
	ab.MoveRight(desiredCol)

	return ab.String()
}
func (p *Prompt) SetInput(val string) {
	p.InputBuffer.Set(val)
}

func (p *Prompt) InputString() string {
	return p.InputBuffer.String()
}

// Disable prev function
func (p *Prompt) Set(bool) {}
