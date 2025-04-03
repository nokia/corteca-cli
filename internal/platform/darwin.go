//go:build darwin
// +build darwin

// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package platform

import (
	"os"
	"path/filepath"
)

const DefaultSSHLog = "/dev/null"

func SetConfigPaths() (string, string) {
	systemConfigRoot := filepath.Join("/", "usr", "local", "etc", "corteca")
	userConfigRoot := filepath.Join(os.Getenv("HOME"), ".config", "corteca")
	return systemConfigRoot, userConfigRoot
}
