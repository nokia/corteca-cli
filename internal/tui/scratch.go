//go:build ignore

// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"github.com/nokia/corteca-cli/internal/tui"
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
