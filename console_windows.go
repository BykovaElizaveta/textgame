//go:build windows

package main

import (
	"os"
	"syscall"
	"unsafe"
)

// STD_OUTPUT_HANDLE в WinAPI равен (DWORD)-11.
const stdOutputHandle = uintptr(0xFFFFFFFB)

const enableVirtualTerminalProcessing = 0x0004

var (
	kernel32DLL        = syscall.NewLazyDLL("kernel32.dll")
	getStdHandleProc   = kernel32DLL.NewProc("GetStdHandle")
	getConsoleModeProc = kernel32DLL.NewProc("GetConsoleMode")
	setConsoleModeProc = kernel32DLL.NewProc("SetConsoleMode")
)

func enableVirtualTerminal(f *os.File) bool {
	if f != os.Stdout && f != os.Stderr {
		return true
	}

	handle, _, _ := getStdHandleProc.Call(stdOutputHandle)
	if handle == 0 || handle == ^uintptr(0) {
		return false
	}

	var mode uint32
	ok, _, _ := getConsoleModeProc.Call(handle, uintptr(unsafe.Pointer(&mode)))
	if ok == 0 {
		return false
	}

	ok, _, _ = setConsoleModeProc.Call(handle, uintptr(mode|enableVirtualTerminalProcessing))
	return ok != 0
}
