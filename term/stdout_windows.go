// +build windows

package term

import (
	"bytes"
	"io"
	"os"
	"unsafe"

	"github.com/gohxs/hsterm/ansi"
)

const (
	COMMON_LVB_UNDERSCORE = 0x8000
)

// Pos hold cursor position on \033[s
type Pos struct {
	x, y int
}

type stdoutWriter struct {
	io.Writer         // Original writer
	fd        uintptr // file descriptor
	err       error
	storedPos Pos
}

//NewStdoutWriter windows Ansi writer
func NewStdoutWriter(w io.Writer) io.Writer {
	var (
		err error
		fd  uintptr
	)
	f, ok := w.(*os.File)
	if !ok {
		err = ErrNotTerminal
	} else {
		fd = uintptr(f.Fd())
	}

	return &stdoutWriter{Writer: w, fd: fd, err: err}
}

func (w *stdoutWriter) Write(b []byte) (written int, err error) {
	if w.err != nil {
		return 0, err
	}

	// Or create a pipe and loop
	br := bytes.NewReader(b) // Read whateever is in b
	ar := ansi.NewReader(br)

	for {
		val, err := ar.ReadEscape()
		if err != nil {
			return 0, err
		}

		if val.Type == ansi.TypeEscape {
			info, err := GetConsoleScreenBufferInfo(w.fd)
			if err != nil {
				return 0, err
			}

			change := true
			//////////////////////////////////////
			// Cursor Postioning
			/////////////////////////////////////
			switch val.Value {
			case "\x1B[s": // Save and restore cursor
				w.storedPos.x = int(info.dwCursorPosition.x)
				w.storedPos.y = int(info.dwCursorPosition.y)
			case "\x1B[u":
				info.dwCursorPosition.x = short(w.storedPos.x)
				info.dwCursorPosition.y = short(w.storedPos.y)
			case "\x1B[H": // Direct set position
				info.dwCursorPosition.x = short(min(lookIndex(val.Attr, 1, 0)-1, 0))
				info.dwCursorPosition.y = short(min(lookIndex(val.Attr, 0, 0)-1, 0))
			case "\x1B[A": // go UP
				info.dwCursorPosition.y -= short(min(lookIndex(val.Attr, 0, 1), 1))
			case "\x1B[B": // go Down
				info.dwCursorPosition.y += short(min(lookIndex(val.Attr, 0, 1), 1))
			case "\x1B[C": // go Right
				info.dwCursorPosition.x += short(min(lookIndex(val.Attr, 0, 1), 1))
			case "\x1B[D": // go UP
				info.dwCursorPosition.x -= short(min(lookIndex(val.Attr, 0, 1), 1))
			default:
				change = false
			}
			if change {
				SetConsoleCursorPosition(w.fd, &info.dwCursorPosition)
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
				color := word(0x7) // Default reset ?
				//intensity := 0

				for _, c := range val.Attr {
					switch {
					case c == 0:
						color = word(0x07)
					case c == 1: // Foregroundintensity
						color |= 0x08
					case c >= 30 && c < 37:
						c -= 30
						bits := ((c & 0x1) << 2) | c&0x2 | ((c & 0x4) >> 2) // swap bit 1 and 3
						color = color&0xFFF8 | word(bits)                   // Invert red blue
					case c >= 40 && c < 47:
						c -= 40
						bits := ((c & 0x1) << 2) | c&0x2 | ((c & 0x4) >> 2) // swap bit 1 and 3
						color = color&0xFF8F | (word(bits << 4))            // Invert red blue
					case c == 4:
						color |= COMMON_LVB_UNDERSCORE | 0x7
					}
				}
				kernel.SetConsoleTextAttribute(w.fd, uintptr(color))
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
					cinfo, err := GetConsoleCursorInfo(w.fd)
					if err != nil {
						return written, err
					}
					cinfo.bVisible = visible

					err = SetConsoleCursorInfo(w.fd, cinfo)
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
	sbi, err := GetConsoleScreenBufferInfo(w.fd)
	if err != nil {
		return err
	}

	var written int
	var size short
	switch mode {
	case 0: // From cursor to end of line
		size = sbi.dwSize.x - sbi.dwCursorPosition.x // length
	case 1: // from line beginning to cursor
		size = sbi.dwCursorPosition.x + 1
		sbi.dwCursorPosition.x = 0
	case 2: // Entire line
		size = sbi.dwSize.x
		sbi.dwCursorPosition.x = 0
	}
	// Do the clear
	kernel.FillConsoleOutputAttribute(w.fd, uintptr(0x7), // Why
		uintptr(size),
		sbi.dwCursorPosition.ptr(),
		uintptr(unsafe.Pointer(&written)),
	)
	kernel.FillConsoleOutputCharacterW(w.fd, uintptr(' '),
		uintptr(size),
		sbi.dwCursorPosition.ptr(),
		uintptr(unsafe.Pointer(&written)),
	)

	return nil

}
func (w *stdoutWriter) EraseDisplay(mode int) error {
	sbi, err := GetConsoleScreenBufferInfo(w.fd)
	if err != nil {
		return err
	}

	var size short
	switch mode {
	case 0: // From cursor to end of screen
		size = (sbi.dwCursorPosition.y - sbi.dwSize.y) * sbi.dwSize.x // Prox lines
		size += sbi.dwCursorPosition.x
	case 1: // From beginning to cursor
		size = sbi.dwCursorPosition.y * sbi.dwSize.x
		size += sbi.dwCursorPosition.x + 1
		sbi.dwCursorPosition.x = 0
		sbi.dwCursorPosition.y = 0
	case 2: // Entire screen
		size = sbi.dwSize.y * sbi.dwSize.x
		sbi.dwCursorPosition.x = 0
		sbi.dwCursorPosition.y = 0
	}
	var written int
	kernel.FillConsoleOutputAttribute(w.fd, uintptr(0x7),
		uintptr(size),
		sbi.dwCursorPosition.ptr(),
		uintptr(unsafe.Pointer(&written)),
	)
	return kernel.FillConsoleOutputCharacterW(w.fd, uintptr(' '),
		uintptr(size),
		sbi.dwCursorPosition.ptr(),
		uintptr(unsafe.Pointer(&written)),
	)

	return ErrNotImplemented
}
