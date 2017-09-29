// +build windows

package termutils

import (
	"errors"
	"io"
	"os"
	"unsafe"
)

// Keys
const (
	VkCancel   = 0x03
	VkBack     = 0x08
	VkTab      = 0x09
	VkReturn   = 0x0D
	VkShift    = 0x10
	VkControl  = 0x11
	VkMenu     = 0x12
	VkEscape   = 0x1B
	VkLeft     = 0x25
	VkUp       = 0x26
	VkRight    = 0x27
	VkDown     = 0x28
	VkDelete   = 0x2E
	VkLShift   = 0xA0
	VkRShift   = 0xA1
	VkLControl = 0xA2
	VkRControl = 0xA3
)

// RawReader translate input record to ANSI escape sequence.
// To provides same behavior as unix terminal.
type stdinReader struct {
	io.Reader
	ctrlKey bool
	altKey  bool
}

// NewStdinReader Creates a reader
func NewStdinReader(rd io.Reader) io.Reader {
	r := &stdinReader{Reader: rd}
	return r
}

// only process one action in one read
func (r *stdinReader) Read(buf []byte) (int, error) {
	ir := new(inputRecord)
	//var read int
	var err error
	stdinFile, ok := r.Reader.(*os.File)
	if !ok {
		return 0, errors.New("Not stdin")
	}
	stdin := stdinFile.Fd()

	for {
		ir, _, err = readConsoleInputW(stdin)
		//_, _, err = syscall.Syscall(procReadConsoleInputW.Addr(), stdin, uintptr(unsafe.Pointer(ir)), 1, uintptr(unsafe.Pointer(&read)))
		if err != nil {
			return 0, err
		}
		if ir.eventType != eventKey {
			continue // Continue
		}
		ker := (*keyEventRecord)(unsafe.Pointer(&ir.event[0]))
		if ker.keyDown == 0 { // keyup
			if r.ctrlKey || r.altKey {
				switch ker.virtualKeyCode {
				case VkRControl, VkLControl:
					r.ctrlKey = false
				case VkMenu: //alt
					r.altKey = false
				}
			}
			continue // Continue
		}

		if ker.unicodeChar == 0 {
			var target string
			switch ker.virtualKeyCode {
			case VkRControl, VkLControl:
				r.ctrlKey = true
			case VkMenu: //alt
				r.altKey = true
			case VkLeft:
				target = "\x1B[D"
			case VkRight:
				target = "\x1B[C"
			case VkUp:
				target = "\x1B[A"
			case VkDown:
				target = "\x1B[B"
			}
			if len(target) != 0 {
				return r.write(buf, target) // Break
			}
			continue // Continue
		}
		char := rune(ker.unicodeChar)
		if r.ctrlKey {
			switch char {
			case 'A':
				char = '\x01'
			case 'E':
				char = '\x05'
			case 'R':
				char = '\x12'
			case 'S':
				char = '\x13'
			}
		} else if r.altKey {
			switch char {
			case VkBack:
				char = '\b'
			}
			return r.writeEsc(buf, char) // Write \x1B AND CHAR break?
		}
		//break here
		return r.write(buf, string(char)) // breal
	}
}

func (r *stdinReader) writeEsc(b []byte, char rune) (int, error) {
	b[0] = '\x1B'
	n := copy(b[1:], []byte(string(char)))
	return n + 1, nil
}

func (r *stdinReader) write(b []byte, char string) (int, error) {
	n := copy(b, []byte(char))
	return n, nil
}

// ANSI WRITER/////////////////
