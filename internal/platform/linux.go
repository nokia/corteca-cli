//go:build linux
// +build linux

// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package platform

import (
	"os"
	"path/filepath"
)

func SetConfigPaths() (string, string) {
	homeDir, _ := os.UserHomeDir()
	systemConfigRoot := filepath.Join("/", "etc", "corteca")
	userConfigRoot := filepath.Join(homeDir, ".config", "corteca")
	return systemConfigRoot, userConfigRoot
}
