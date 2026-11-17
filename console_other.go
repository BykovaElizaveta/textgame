//go:build !windows

package main

import "os"

func enableVirtualTerminal(_ *os.File) bool {
	return true
}
