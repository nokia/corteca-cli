//go:build ignore

package main

import (
	"corteca/internal/tui"
	"time"
)

func main() {
	prog := tui.PromptForProgress("Testing bar")
	defer close(prog)
	for i := 1; i <= 15; i++ {
		prog <- tui.ProgressUpdate{Current: int64(i), Total: 15}
		time.Sleep(250 * time.Millisecond)
	}
}
