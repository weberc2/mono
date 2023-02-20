package main

import (
	"os"

	"golang.org/x/sys/unix"
)

func isTTY() bool {
	_, err := unix.IoctlGetWinsize(int(os.Stdout.Fd()), unix.TIOCGWINSZ)
	return err == nil
}
