// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package cmd

import (
	"corteca/internal/configuration"
	"corteca/internal/toolchain"
	"corteca/internal/tui"
	"fmt"

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

func doBuildApp(selectedName string) {
	requireProjectContext()

	if selectedName == "" {
		selectedName = config.Build.Default
	} else if _, ok := config.Build.Toolchains[selectedName]; !ok {
		failOperation(fmt.Sprintf("no configuration present for toolchain: \"%v\"", selectedName))
	}

	for toolchainName, toolchain := range config.Build.Toolchains {

		if !buildAll && selectedName != toolchainName {
			continue
		}

		fmt.Printf("Building with toolchain '%s'...\n", toolchainName)
		err := doBuildTarget(toolchain.Image, toolchain.ConfigFile, config.Build.Options)
		if err != nil {
			tui.DisplayErrorMsg(fmt.Sprintf("Error building '%s': %s", toolchainName, err.Error()))
			failOperation("Build failed")
		}
		tui.DisplaySuccessMsg(fmt.Sprintf("Application '%v' was built successfully with toolchain '%v'\n", config.App.Name, toolchainName))
	}
}

func doBuildTarget(targetImage string, customConfig string, buildOptions configuration.BuildOptions) error {
	err := toolchain.Invoke(targetImage, projectRoot, customConfig, buildOptions)

	if err != nil {
		return fmt.Errorf("invoking toolchain image '%v': %v", targetImage, err)
	}

	return nil
}
