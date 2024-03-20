// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package tui

import (
	"errors"

	"github.com/pterm/pterm"
	"github.com/spf13/afero"
)

var ErrInputCancelled = errors.New("input cancelled")

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
	pterm.Success.Println(msg)
}

func DisplayErrorMsg(msg string) {
	pterm.Error.Println(msg)
}
