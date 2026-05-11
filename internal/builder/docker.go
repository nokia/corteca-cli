// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package builder

import (
	"github.com/nokia/corteca-cli/internal/configuration"
	"github.com/nokia/corteca-cli/internal/fsutil"
	"github.com/nokia/corteca-cli/internal/tui"
	"fmt"
	"os"
	"os/exec"
)

func execDocker(args ...string) error {
	tui.SetOutputColor(tui.CBlue, os.Stderr)
	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	tui.ResetOutputColor(os.Stdout)
	return err
}

func enableCrossCompilation(crossCompileSettings configuration.CrossCompileConfig) error {
	args := []string{"run", "--rm", "--privileged", crossCompileSettings.Image}
	args = append(args, crossCompileSettings.Args...)

	if err := execDocker(args...); err != nil {
		return fmt.Errorf("failed to setup cross-compilation: %v", err)
	}
	return nil
}

func prepareDockerBuildArgs(platform string, buildOptions configuration.BuildOptions, appDir, distPath string) ([]string, error) {
	args := []string{"buildx", "build", "--platform", platform}

	if !buildOptions.SkipHostEnv {
		inheritEnvironment(&args)
	}

	if tui.DisableColoredOutput {
		args = append(args, "--progress=plain")
	}
	args = append(args, "-f", "Dockerfile")

	if err := fsutil.CleanupOrCreateFolder(distPath); err != nil {
		return nil, fmt.Errorf("failed to prepare dist directory: %v", err)
	}
	args = append(args, "--output", "type=local,dest="+distPath)

	args = append(args, appDir)
	return args, nil
}
