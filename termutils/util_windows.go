// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

// Package terminal provides support functions for dealing with terminals, as
// commonly found on UNIX systems.
//
// Putting a terminal into raw mode is the most common requirement:
//
// 	oldState, err := terminal.MakeRaw(0)
// 	if err != nil {
// 	        panic(err)
// 	}
// 	defer terminal.Restore(0, oldState)
package termutils

import (
	"syscall"
	"unsafe"
)

const (
	eventKey              = 0x0001
	eventMouse            = 0x0002
	eventWindowBufferSize = 0x0004
	eventMenu             = 0x0008
	eventFocus            = 0x0010
)

const (
	enableLineInput       = 2
	enableEchoInput       = 4
	enableProcessedInput  = 1
	enableWindowInput     = 8
	enableMouseInput      = 16
	enableInsertMode      = 32
	enableQuickEditMode   = 64
	enableExtendedFlags   = 128
	enableAutoPosition    = 256
	enableProcessedOutput = 1
	enableWrapAtEolOutput = 2
)

var kernel32 = syscall.NewLazyDLL("kernel32.dll")

var (
	procGetConsoleMode              = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode              = kernel32.NewProc("SetConsoleMode")
	procGetConsoleScreenBufferInfo  = kernel32.NewProc("GetConsoleScreenBufferInfo")
	procGetConsoleCursorInfo        = kernel32.NewProc("GetConsoleCursorInfo")
	procSetConsoleCursorInfo        = kernel32.NewProc("SetConsoleCursorInfo")
	procSetConsoleCursorPosition    = kernel32.NewProc("SetConsoleCursorPosition")
	procSetConsoleTextAttribute     = kernel32.NewProc("SetConsoleTextAttribute")
	procFillConsoleOutputCharacterW = kernel32.NewProc("FillConsoleOutputCharacterW")
	procFillConsoleOutputAttribute  = kernel32.NewProc("FillConsoleOutputAttribute")
	procReadConsoleInputW           = kernel32.NewProc("ReadConsoleInputW")
)

type (
	short int16
	word  uint16
	dword uint32
	wchar uint16

	coord struct {
		x short
		y short
	}
	smallRect struct {
		left   short
		top    short
		right  short
		bottom short
	}
	consoleScreenBufferInfo struct {
		size              coord
		cursorPosition    coord
		attributes        word
		window            smallRect
		maximumWindowSize coord
	}
	keyEventRecord struct {
		keyDown         int32
		repeatCount     word
		virtualKeyCode  word
		virtualScanCode word
		unicodeChar     wchar
		controlKeyState dword
	}
	inputRecord struct {
		eventType word
		padding   uint16
		event     [16]byte
	}
	consoleCursorInfo struct {
		size    dword
		visible bool
	}
)

type State struct {
	Mode uint32
}

// IsTerminal returns true if the given file descriptor is a terminal.
func IsTerminal(fd int) bool {
	var st uint32
	r, _, e := syscall.Syscall(procGetConsoleMode.Addr(), 2, uintptr(fd), uintptr(unsafe.Pointer(&st)), 0)
	return r != 0 && e == 0
}

// MakeRaw put the terminal connected to the given file descriptor into raw
// mode and returns the previous state of the terminal so that it can be
// restored.
func MakeRaw(fd int) (*State, error) {
	var st uint32
	_, _, e := syscall.Syscall(procGetConsoleMode.Addr(), 2, uintptr(fd), uintptr(unsafe.Pointer(&st)), 0)
	if e != 0 {
		return nil, error(e)
	}
	raw := st &^ (enableEchoInput | enableProcessedInput | enableLineInput | enableProcessedOutput)
	_, _, e = syscall.Syscall(procSetConsoleMode.Addr(), 2, uintptr(fd), uintptr(raw), 0)
	if e != 0 {
		return nil, error(e)
	}
	return &State{st}, nil
}

// GetState returns the current state of a terminal which may be useful to
// restore the terminal after a signal.
func GetState(fd int) (*State, error) {
	var st uint32
	_, _, e := syscall.Syscall(procGetConsoleMode.Addr(), 2, uintptr(fd), uintptr(unsafe.Pointer(&st)), 0)
	if e != 0 {
		return nil, error(e)
	}
	return &State{st}, nil
}

// Restore restores the terminal connected to the given file descriptor to a
// previous state.
func Restore(fd int, state *State) error {
	_, _, e := syscall.Syscall(procSetConsoleMode.Addr(), 2, uintptr(fd), uintptr(state.Mode), 0)
	if e != 0 {
		return error(e)
	}
	return nil
}

// GetSize returns the dimensions of the given terminal.
func GetSize(fd int) (width, height int, err error) {
	var info consoleScreenBufferInfo
	_, _, e := syscall.Syscall(procGetConsoleScreenBufferInfo.Addr(), 2, uintptr(fd), uintptr(unsafe.Pointer(&info)), 0)
	if e != 0 {
		return 0, 0, error(e)
	}
	return int(info.size.x), int(info.size.y), nil
}

// Private helpers
func getConsoleScreenBufferInfo(fd uintptr) (*consoleScreenBufferInfo, error) {
	var info consoleScreenBufferInfo
	_, _, e := syscall.Syscall(procGetConsoleScreenBufferInfo.Addr(), 2, uintptr(fd), uintptr(unsafe.Pointer(&info)), 0)
	if e != 0 {
		return nil, error(e)
	}
	return &info, nil
}

func readConsoleInputW(fd uintptr) (*inputRecord, int, error) {
	var ir inputRecord
	var read int
	_, _, e := syscall.Syscall6(procReadConsoleInputW.Addr(), 4,
		fd,
		uintptr(unsafe.Pointer(&ir)), 1, uintptr(unsafe.Pointer(&read)),
		0, 0,
	)
	if e != 0 {
		return nil, 0, error(e)
	}
	return &ir, read, nil
}

func getConsoleCursorInfo(fd uintptr) (*consoleCursorInfo, error) {
	var info consoleCursorInfo
	_, _, e := syscall.Syscall(procGetConsoleCursorInfo.Addr(), 2, uintptr(fd), uintptr(unsafe.Pointer(&info)), 0)
	if e != 0 {
		return nil, error(e)
	}
	return &info, nil
}

func setConsoleCursorInfo(fd uintptr, p *consoleCursorInfo) error {
	_, _, e := syscall.Syscall(procSetConsoleCursorInfo.Addr(), 2, uintptr(fd), uintptr(unsafe.Pointer(p)), 0)
	if e != 0 {
		return error(e)
	}
	return nil
}

func setConsoleCursorPosition(fd uintptr, c *coord) error {
	_, _, e := syscall.Syscall(procSetConsoleCursorPosition.Addr(), 2,
		uintptr(fd),
		uintptr(*(*int32)(unsafe.Pointer(c))), // weird
		0,
	)
	if e != 0 {
		return error(e)
	}
	return nil
}

func setConsoleTextAttribute(fd uintptr, color uintptr) error {
	_, _, e := syscall.Syscall(procSetConsoleTextAttribute.Addr(), 2,
		uintptr(fd),
		uintptr(color),
		0,
	)
	if e != 0 {
		return error(e)
	}
	return nil
}
func fillConsoleOutputAttribute(fd uintptr, color uintptr, size uintptr, cursor *coord, written *int) error {
	_, _, e := syscall.Syscall6(procFillConsoleOutputAttribute.Addr(), 5,
		fd,
		color, size, uintptr(*(*int32)(unsafe.Pointer(cursor))),
		uintptr(unsafe.Pointer(written)), 0)
	if e != 0 {
		return error(e)
	}
	return nil
}
func fillConsoleOutputCharacterW(fd uintptr, char uintptr, size uintptr, cursor *coord, written *int) error {
	_, _, e := syscall.Syscall6(procFillConsoleOutputCharacterW.Addr(), 5,
		fd,
		char, size, uintptr(*(*int32)(unsafe.Pointer(cursor))),
		uintptr(unsafe.Pointer(written)), 0)
	if e != 0 {
		return error(e)
	}
	return nil
}
