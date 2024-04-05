//go:build windows
// +build windows

// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package platform

import (
	"os"
	"path/filepath"
)

func SetConfigPaths() (string, string) {
	systemConfigRoot := filepath.Join(os.Getenv("PROGRAMDATA"), "Corteca")
	userConfigRoot := filepath.Join(os.Getenv("APPDATA"), "Corteca")
	return systemConfigRoot, userConfigRoot
}
