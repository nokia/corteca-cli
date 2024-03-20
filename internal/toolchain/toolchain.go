// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package toolchain

import (
	"fmt"
	"os"
	"os/exec"
)

const (
	csReset    = "\033[0m"
	sBold      = "\033[1m"
	sUnderline = "\033[4m"
	sStrike    = "\033[9m"
	sItalic    = "\033[3m"

	cRed    = "\033[31m"
	cGreen  = "\033[32m"
	cYellow = "\033[33m"
	cBlue   = "\033[34m"
	cPurple = "\033[35m"
	cCyan   = "\033[36m"
	cWhite  = "\033[37m"
)

func Invoke(imageName, appDir, configFile string) error {
	if len(configFile) > 0 {
		return execDocker("run",
			"--rm",
			"-v",
			configFile+":/buildroot/.config",
			"-e",
			"REBUILD_BUILDROOT=yes",
			"-v", fmt.Sprintf("%v:/app", appDir),
			imageName)
	} else {
		return execDocker("run",
			"--rm",
			"-v", fmt.Sprintf("%v:/app", appDir),
			imageName)
	}
}

func execDocker(args ...string) error {
	// TODO: check if stdout is a tty first
	fmt.Fprint(os.Stdout, cBlue)
	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	fmt.Fprint(os.Stdout, csReset)
	return err
}
