//go:build windows

package main

import "os"

// openTTY opens the terminal device for direct TUI output,
// bypassing stdout/stderr so they remain clean for piping.
func openTTY() (*os.File, error) {
	return os.OpenFile("CONOUT$", os.O_WRONLY, 0)
}
