// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package tui

import (
	"errors"
	"fmt"
	"os"

	"github.com/pterm/pterm"
	"github.com/spf13/afero"
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
	if DisableColoredOutput || !(term.IsTerminal(int(os.Stdout.Fd())) || term.IsTerminal(int(os.Stderr.Fd()))) {
		DisableColoredOutput = true
		pterm.DisableColor()
	}
}

func LogNormal(format string, args ...any) {
	format += "\n"
	fmt.Fprintf(os.Stdout, format, args...)
}

func LogError(format string, args ...any) {
	SetOutputColor(CRed)
	format += "\n"
	fmt.Fprintf(os.Stderr, format, args...)
	ResetOutputColor()
}

func SetOutputColor(colorSeq string) {
	if !DisableColoredOutput {
		fmt.Fprintf(os.Stderr, colorSeq)
	}
}

func ResetOutputColor() {
	if !DisableColoredOutput {
		fmt.Fprintf(os.Stderr, CsReset)
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

func DisplayHelpMsg(msg string) {
	pterm.ThemeDefault.InfoMessageStyle.Println(msg)
}

type ProgressBar struct {
	file    afero.File
	progBar *pterm.ProgressbarPrinter
}

func (pb *ProgressBar) Read(p []byte) (int, error) {
	n, err := pb.file.Read(p)

	if err != nil {
		return n, err
	}

	pb.progBar.Add(n)

	return n, nil
}

func (pb *ProgressBar) Close() {
	pb.progBar.Stop()
}

func PromptForProgress(f afero.File, label string) (*ProgressBar, error) {

	fileInfo, err := f.Stat()

	if err != nil {
		return nil, err
	}

	pb := ProgressBar{
		file:    f,
		progBar: pterm.DefaultProgressbar.WithTotal(int(fileInfo.Size())).WithTitle(label).WithMaxWidth(-1).WithCurrent(0),
	}

	pb.progBar, err = pb.progBar.Start()
	if err != nil {
		return nil, err
	}

	return &pb, nil
}

func DisplaySuccessMsg(msg string) {
	SetOutputColor(CGreen)
	LogNormal(msg)
	ResetOutputColor()
}

func DisplayErrorMsg(msg string) {
	LogError(msg)
}
