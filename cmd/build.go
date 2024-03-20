// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package cmd

import (
	"corteca/internal/toolchain"
	"corteca/internal/tui"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build [TOOLCHAIN]",
	Short: "Build application",
	Long:  `Build the application using the appropriate build toolchain`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		toolchain := ""
		if len(args) > 0 {
			toolchain = args[0]
		}

		doBuildApp(toolchain)
	},
}

var buildAll bool

func init() {
	buildCmd.PersistentFlags().BoolVar(&buildAll, "all", false, "Build for all platforms")
	rootCmd.AddCommand(buildCmd)
}

func AbsPath(path, target string) (string, error) {
	if filepath.IsAbs(path) {
		return filepath.Clean(path), nil
	}
	return filepath.Abs(filepath.Join(target, path))
}

func doBuildApp(toolchain string) {
	requireProjectContext()

	if toolchain == "" {
		toolchain = config.Toolchain.Default
	} else if _, ok := config.Toolchain.Targets[toolchain]; !ok {
		failOperation(fmt.Sprintf("no configuration present for toolchain: \"%v\"", toolchain))
	}

	for toolchainName := range config.Toolchain.Targets {

		if buildAll || toolchain == toolchainName {
			fmt.Printf("\n==> Processing toolchain '%v'...\n", toolchainName)

			toolchainImage, customConfig, err := doSetupBuild(toolchainName)
			if err != nil {
				processBuildError(toolchainName, err)
				continue
			}

			fmt.Printf("Building with toolchain '%v', using toolchain image '%v' and custom config '%v'...\n", toolchainName, toolchainImage, customConfig)
			err = doBuildTarget(toolchainImage, customConfig)
			if err != nil {
				processBuildError(toolchainName, err)
				continue
			}

			tui.DisplaySuccessMsg(fmt.Sprintf("Application '%v' was built successfully with toolchain '%v'\n", config.App.Name, toolchainName))
		}
	}
}

func doSetupBuild(toolchainName string) (targetImage string, customConfig string, err error) {
	targetImage = config.Toolchain.Targets[toolchainName].Image
	if len(config.Toolchain.Configs[toolchainName].Config) > 0 {
		customConfig, err = AbsPath(config.Toolchain.Configs[toolchainName].Config, projectRoot)

		if err != nil {
			return "", "", fmt.Errorf("retrieving custom config for toolchain '%v': %v", toolchainName, err)
		}
	}

	return
}

func doBuildTarget(targetImage string, customConfig string) error {
	err := toolchain.Invoke(targetImage, projectRoot, customConfig)

	if err != nil {
		return fmt.Errorf("invoking toolchain image '%v': %v", targetImage, err)
	}

	return nil
}

func processBuildError(toolchainName string, err error) {
	fmt.Fprintf(os.Stderr, "Error while: %v\n", err)
	tui.DisplayErrorMsg(fmt.Sprintf("Aborting target '%v'\n", toolchainName))
}
