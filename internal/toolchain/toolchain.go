// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package toolchain

import (
	"corteca/internal/configuration"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

var InheritEnvironmentVars = []string{
	"HTTP_PROXY",
	"http_proxy",
	"HTTPS_PROXY",
	"https_proxy",
	"FTP_PROXY",
	"ftp_proxy",
	"NO_PROXY",
	"no_proxy",
}

func absPath(path, target string) (string, error) {
	if filepath.IsAbs(path) {
		return filepath.Clean(path), nil
	}
	return filepath.Abs(filepath.Join(target, path))
}

func Invoke(imageName, appDir, configFile string, buildOptions configuration.BuildOptions) error {
	args := []string{
		"run",
		"--rm",
		"-v", fmt.Sprintf("%v:/app", appDir),
	}
	if !buildOptions.SkipHostEnv {
		inheritEnvironment(&args)
	}
	for key, value := range buildOptions.Env {
		args = append(args, "--env", fmt.Sprintf("%s=%s", key, value))
	}
	if len(configFile) > 0 {
		configFilePath, err := absPath(configFile, appDir)
		if err != nil {
			return err
		}
		args = append(args,
			"-v",
			configFilePath+":/buildroot/.config",
			"-e",
			"REBUILD_BUILDROOT=yes")
	}
	args = append(args, imageName)
	return execDocker(args...)
}

func inheritEnvironment(args *[]string) {
	for _, varName := range InheritEnvironmentVars {
		varValue := os.Getenv(varName)
		if varValue != "" {
			(*args) = append(*args, "--env", fmt.Sprintf("%s=%s", varName, varValue))
		}
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
