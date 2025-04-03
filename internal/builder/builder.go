// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package builder

import (
	"corteca/internal/configuration"
	"corteca/internal/fsutil"
	"fmt"
	"os"
)

const (
	buildArg = "--build-arg"
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

func inheritEnvironment(args *[]string) {
	for _, varName := range InheritEnvironmentVars {
		varValue := os.Getenv(varName)
		if varValue != "" {
			(*args) = append(*args, buildArg, fmt.Sprintf("%s=%s", varName, varValue))
		}
	}
}

func BuildRootFS(appDir, tmpBuildPath string, settings configuration.ArchitectureSettings, appSettings configuration.AppSettings, buildSettings configuration.BuildSettings) error {

	if err := fsutil.EnsureDirExists(tmpBuildPath); err != nil {
		return fmt.Errorf("failed to create dist directory: %v", err)
	}

	if err := enableCrossCompilation(buildSettings.CrossCompile); err != nil {
		return err
	}

	args, err := prepareDockerBuildArgs(settings.Platform, buildSettings.Options, appDir, tmpBuildPath)
	if err != nil {
		return err
	}

	if err := execDocker(args...); err != nil {
		return fmt.Errorf("docker build failed: %v", err)
	}

	return nil
}
