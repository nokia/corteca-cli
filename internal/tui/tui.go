// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package tui

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/pterm/pterm"
	"golang.org/x/term"
)

const (
	CsReset    = "\033[0m"
	SBold      = "\033[1m"
	SUnderline = "\033[4m"
	SStrike    = "\033[9m"
	SItalic    = "\033[3m"

	CRed    = "\033[31m"
	CGreen  = "\033[32m"
	CYellow = "\033[33m"
	CBlue   = "\033[34m"
	CPurple = "\033[35m"
	CCyan   = "\033[36m"
	CWhite  = "\033[37m"
)

var (
	ErrInputCancelled    = errors.New("input cancelled")
	DisableColoredOutput bool
)

// If no-color flag is set or no terminal found then remove colored output
func DefineOutputColor() {
	if DisableColoredOutput || !(term.IsTerminal(int(os.Stdout.Fd())) && term.IsTerminal(int(os.Stderr.Fd()))) {
		DisableColoredOutput = true
		pterm.DisableColor()
	}
}

func LogNormal(format string, args ...any) {
	format += "\n"
	fmt.Fprintf(os.Stderr, format, args...)
}

func LogError(format string, args ...any) {
	SetOutputColor(CRed, os.Stderr)
	format += "\n"
	fmt.Fprintf(os.Stderr, format, args...)
	ResetOutputColor(os.Stderr)
}

func SetOutputColor(colorSeq string, out io.Writer) {
	if !DisableColoredOutput {
		fmt.Fprintf(out, colorSeq)
	}
}

func ResetOutputColor(out io.Writer) {
	if !DisableColoredOutput {
		fmt.Fprintf(out, CsReset)
	}
}

func PromptForValue(label string, defaultValue string) (string, error) {
	result, err := pterm.DefaultInteractiveTextInput.WithDefaultValue(defaultValue).Show(label)
	if result == "" && err == nil {
		return defaultValue, nil
	}
	return result, err
}

func PromptForSelection(label string, items []string, defaultValue string) (string, error) {
	result, err := pterm.DefaultInteractiveSelect.WithOptions(items).WithDefaultOption(defaultValue).Show(label)
	return result, err
}

func PromptForConfirm(label string, defaultYes bool) (bool, error) {
	result, err := pterm.DefaultInteractiveConfirm.WithDefaultValue(defaultYes).Show(label)
	return result, err
}

func PromptForPassword(label string) (string, error) {
	result, err := pterm.DefaultInteractiveTextInput.WithMask("*").Show(label)
	return result, err
}

type ProgressUpdate struct {
	Current int64
	Total   int64
}

func PromptForProgress(label string) chan<- ProgressUpdate {
	const max = 100
	ch := make(chan ProgressUpdate, 8)
	bar, _ := pterm.DefaultProgressbar.
		WithTitle(label).
		WithTotal(max).
		WithShowCount(false).
		WithShowElapsedTime(false).
		Start()
	go func() {
		for update := range ch {
			current := int((update.Current * int64(max)) / update.Total)
			diff := current - bar.Current
			bar.Add(diff)
		}
		bar.Stop()
	}()
	return ch
}

func DisplayHelpMsg(msg string) {
	pterm.ThemeDefault.InfoMessageStyle.Println(msg)
}

func DisplaySuccessMsg(msg string) {
	SetOutputColor(CGreen, os.Stderr)
	LogNormal(msg)
	ResetOutputColor(os.Stderr)
}

func DisplayErrorMsg(msg string) {
	LogError(msg)
}

func LogOutData(format string, args ...any) {
	format += "\n"
	fmt.Fprintf(os.Stdout, format, args...)
}
