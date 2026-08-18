//go:build windows

package main

import (
	"os"
	"syscall"
	"unsafe"
)

const enableVirtualTerminalProcessing = 0x0004

func enableVirtualTerminal() {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	setConsoleMode := kernel32.NewProc("SetConsoleMode")
	getConsoleMode := kernel32.NewProc("GetConsoleMode")
	setConsoleOutputCP := kernel32.NewProc("SetConsoleOutputCP")

	setConsoleOutputCP.Call(uintptr(65001))

	handle := syscall.Handle(os.Stdout.Fd())
	var mode uint32
	if ret, _, _ := getConsoleMode.Call(uintptr(handle), uintptr(unsafe.Pointer(&mode))); ret != 0 {
		mode |= enableVirtualTerminalProcessing
		setConsoleMode.Call(uintptr(handle), uintptr(mode))
	}
}
