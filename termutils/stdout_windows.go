// This area is a TerminalWriter

// +build windows

package termutils

import (
	"bytes"
	"io"
	"os"

	"github.com/gohxs/termu/ansi"
)

// Missing ANSI/VT
//\033[E -- \033[5E   Move to first position of 5th line down
//\033[F -- \033[5F   Move to first position of 5th line previous
//\033[G -- \033[40G  Move to column 40 of current line
//\033[M -- \033[2M   Delete 2 lines if currently in scrolling region
//
//\033[S -- \033[3S   Move everything up 3 lines, bring in 3 new lines
//\033[T -- \033[4T   Scroll down 4, bring previous lines back into view

//\033[7m--  Background, foreground invert
//
const (
	COMMON_LVB_UNDERSCORE = 0x8000
)

type Pos struct { // cursor?
	x, y int
}

type stdoutWriter struct {
	io.Writer         // Original writer
	fd        uintptr // file descriptor
	storedPos Pos
	lastColor word
}

//NewStdoutWriter windows Ansi writer
func NewStdoutWriter(w io.Writer) (io.Writer, error) {
	var fd int

	f, ok := w.(*os.File)
	if !ok {
		return nil, ErrNotTerminal // should fail?
	}

	fd = int(f.Fd())
	if !IsTerminal(fd) {
		return nil, ErrNotTerminal
	}

	// Lets emulate always for testing
	return &stdoutWriter{Writer: w, fd: uintptr(fd), lastColor: 0x7}, nil // Emulated

	st, err := GetState(fd)
	if err != nil {
		return nil, err
	}

	st.Mode |= 0x0004 // Enable VT output // new windows versions
	err = Restore(fd, st)
	if err != nil {
		return &stdoutWriter{Writer: w, fd: uintptr(fd)}, nil // Emulated
	}
	return w, nil // Just passtrough

}

func (w *stdoutWriter) Write(b []byte) (written int, err error) {

	// Or create a pipe and loop
	br := bytes.NewReader(b) // Read whateever is in b
	ar := ansi.NewScanner(br)

	for {
		val, err := ar.Scan()
		if err != nil {
			return 0, err
		}

		if val.Type == ansi.TypeEscape {
			info, err := getConsoleScreenBufferInfo(w.fd)
			if err != nil {
				return 0, err
			}

			change := true
			//////////////////////////////////////
			// Cursor Postioning
			/////////////////////////////////////

			switch val.Value {
			case "\x1B[s": // Save and restore cursor
				w.storedPos.x = int(info.cursorPosition.x)
				w.storedPos.y = int(info.cursorPosition.y)
			case "\x1B[u":
				info.cursorPosition.x = short(w.storedPos.x)
				info.cursorPosition.y = short(w.storedPos.y)
			case "\x1B[H": // Direct set position
				info.cursorPosition.x = short(min(lookIndex(val.Attr, 1, 0)-1, 0))
				info.cursorPosition.y = short(min(lookIndex(val.Attr, 0, 0)-1, 0))
			case "\x1B[A": // go UP
				info.cursorPosition.y -= short(min(lookIndex(val.Attr, 0, 1), 1))
			case "\x1B[B": // go Down
				info.cursorPosition.y += short(min(lookIndex(val.Attr, 0, 1), 1))
			case "\x1B[C": // go Right
				info.cursorPosition.x += short(min(lookIndex(val.Attr, 0, 1), 1))
			case "\x1B[D": // go UP
				info.cursorPosition.x -= short(min(lookIndex(val.Attr, 0, 1), 1))
			default:
				change = false
			}
			if change {
				setConsoleCursorPosition(w.fd, &info.cursorPosition)
				continue // next
			}
			///////////////////////////////////////////////////////////////
			// CLEANERS/STYLING line cleaner, display cleaner, char styling
			////////////////////
			change = true
			switch val.Value { // Cleaners
			case "\x1B[K": // EL Erase in line
				_ = w.EraseLine(lookIndex(val.Attr, 0, 0))
				// Ignore error for now
			case "\x1B[J": // ED Erase in Display
				_ = w.EraseDisplay(lookIndex(val.Attr, 0, 0))
			case "\x1B[m":
				//color := word(0) // color 0?
				color := w.lastColor // Default reset ?
				//intensity := 0

				if len(val.Attr) == 0 { // same as reset "\033[m"
					color = word(0x07)
				}
				attrs := val.Attr
				for ; len(attrs) > 0; attrs = attrs[1:] {
					c := attrs[0]
					switch {
					case c == 0:
						color = word(0x07) // Reset
					case c == 1: // Foregroundintensity and maybe bold
						color |= 0x08
					case c == 4:
						color |= COMMON_LVB_UNDERSCORE | 0x7
					case c == 7: // Swap bits
						tmpFg := color & 0x7       // 3 bits
						tmpBg := color & 0x70 >> 4 // 3 bits

						color = color&0XFFF8 | word(tmpBg)
						color = color&0XFF8F | word(tmpFg<<4)
					case c >= 30 && c <= 37:
						c -= 30
						bits := ((c & 0x1) << 2) | c&0x2 | ((c & 0x4) >> 2) // swap bit 1 and 3
						color = color&0xFFF8 | word(bits)
					case c >= 40 && c <= 47:
						c -= 40
						bits := ((c & 0x1) << 2) | c&0x2 | ((c & 0x4) >> 2) // swap bit 1 and 3
						color = color&0xFF8F | (word(bits << 4))
					case c == 38: // Next should be 5
						n := lookIndex(attrs, 1, -1)
						c256 := lookIndex(attrs, 2, -1)
						if n >= 0 && c256 >= 0 {
							attrs = attrs[2:]
						}
						// Translate 256 to 16 here
					}
				}
				w.lastColor = color // save color state
				setConsoleTextAttribute(w.fd, uintptr(color))
			case "\x1B[?l", "\x1B[?h": // Feature on, feature off
				if len(val.Attr) == 0 {
					break
				}
				fn := val.Attr[0]
				switch fn {
				case 25: // Cursor visibility
					visible := true
					if val.Value[3] == 'l' {
						visible = false
					} else if val.Value[3] == 'h' {
						visible = true
					}
					cinfo, err := getConsoleCursorInfo(w.fd)
					if err != nil {
						return written, err
					}
					cinfo.visible = visible

					err = setConsoleCursorInfo(w.fd, cinfo)
					if err != nil {
						return written, err
					}

				}

			default:
				change = false
			}
			if !change {
				n, err := w.Writer.Write([]byte("^[" + val.Value[1:]))
				written += n
				if err != nil {
					return written, err
				}
			}

			// Write to stdout
		} else if val.Type == ansi.TypeRune {
			// Just write
			n, err := w.Writer.Write([]byte(val.Value))
			written += n
			if err != nil {
				return written, err
			}
		}
	}
}

func (w *stdoutWriter) EraseLine(mode int) error {
	sbi, err := getConsoleScreenBufferInfo(w.fd)
	if err != nil {
		return err
	}

	var written int
	var size short
	switch mode {
	case 0: // From cursor to end of line
		size = sbi.size.x - sbi.cursorPosition.x // length
	case 1: // from line beginning to cursor
		size = sbi.cursorPosition.x + 1
		sbi.cursorPosition.x = 0
	case 2: // Entire line
		size = sbi.size.x
		sbi.cursorPosition.x = 0
	}
	// Do the clear
	fillConsoleOutputAttribute(w.fd, uintptr(w.lastColor),
		uintptr(size),
		&sbi.cursorPosition,
		&written,
	)

	return fillConsoleOutputCharacterW(w.fd, uintptr(' '),
		uintptr(size),
		&sbi.cursorPosition,
		&written,
	)

}
func (w *stdoutWriter) EraseDisplay(mode int) error {
	sbi, err := getConsoleScreenBufferInfo(w.fd)
	if err != nil {
		return err
	}

	var size short
	switch mode {
	case 0: // From cursor to end of screen
		size = (sbi.cursorPosition.y - sbi.size.y) * sbi.size.x // Prox lines
		size += sbi.cursorPosition.x
	case 1: // From beginning to cursor
		size = sbi.cursorPosition.y * sbi.size.x
		size += sbi.cursorPosition.x + 1
		sbi.cursorPosition.x = 0
		sbi.cursorPosition.y = 0
	case 2: // Entire screen
		size = sbi.size.y * sbi.size.x
		sbi.cursorPosition.x = 0
		sbi.cursorPosition.y = 0
	}
	var written int
	fillConsoleOutputAttribute(w.fd, uintptr(w.lastColor),
		uintptr(size),
		&sbi.cursorPosition,
		&written,
	)
	return fillConsoleOutputCharacterW(w.fd, uintptr(' '),
		uintptr(size),
		&sbi.cursorPosition,
		&written,
	)

}
