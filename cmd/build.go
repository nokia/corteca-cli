// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package cmd

import (
	"corteca/internal/builder"
	"corteca/internal/configuration"
	"corteca/internal/tui"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:               "build [TOOLCHAIN]",
	Short:             "Build application",
	Long:              `Build the application using the appropriate build toolchain`,
	Args:              cobra.MaximumNArgs(1),
	ValidArgsFunction: validBuildArgsFunc,
	Run: func(cmd *cobra.Command, args []string) {
		toolchain := ""
		if len(args) > 0 {
			toolchain = args[0]
		}

		doBuildApp(toolchain)
	},
}

var (
	buildAll   bool
	outputType string
)

func init() {
	buildCmd.PersistentFlags().BoolVar(&buildAll, "all", false, "Build for all platforms")
	buildCmd.PersistentFlags().BoolVar(&noRegen, "no-regen", false, "Skip regeneration of templates")
	rootCmd.AddCommand(buildCmd)
}

func doBuildApp(selectedName string) {
	requireProjectContext()
	if !noRegen {
		doRegenTemplates(projectRoot)
	}
	selectedName = getToolchainName(selectedName)
	outputType = strings.ToLower(config.Build.Options.OutputType)
	cmdContext.Toolchain.Image = config.Build.Toolchains.Image
	enableCrossCompilation(config.Build.CrossCompile)

	if buildAll {
		handleBuildAll()
	} else {
		handleSinglePlatformBuild(selectedName)
	}
}

func handleBuildAll() {
	switch outputType {
	case "rootfs":
		buildAllRootfs()
	case "docker", "oci":
		buildMultiPlatformImage()
	default:
		failOperation(fmt.Sprintf("Unknown image type: %s", config.Build.Options.OutputType))
	}
}

func handleSinglePlatformBuild(selectedName string) {
	switch outputType {
	case "rootfs", "docker", "oci":
		buildSinglePlatform(selectedName)
	default:
		failOperation(fmt.Sprintf("Unknown image type: %s", config.Build.Options.OutputType))
	}
}

func getToolchainName(selectedName string) string {
	if selectedName == "" {
		return config.Build.Default
	} else if _, ok := config.Build.Toolchains.Architectures[selectedName]; !ok {
		failOperation(fmt.Sprintf("no configuration present for toolchain: \"%v\"", selectedName))
	}
	return selectedName
}

func buildAllRootfs() {
	for name, settings := range config.Build.Toolchains.Architectures {
		cmdContext.Toolchain.Name = name
		cmdContext.Toolchain.Platform = settings.Platform

		fmt.Printf("Building rootfs with toolchain '%s'...\n", name)
		err := doBuildTarget(name, settings.Platform, config.Build.Options)
		if err != nil {
			handleBuildError(name, err)
		}
		tui.DisplaySuccessMsg(fmt.Sprintf("Application '%v' was built successfully with toolchain '%v'\n", config.App.Name, name))
	}
}

func buildMultiPlatformImage() {
	var platforms []string
	for _, settings := range config.Build.Toolchains.Architectures {
		platforms = append(platforms, settings.Platform)
	}
	platformArg := strings.Join(platforms, ",")

	fmt.Println("Building multi-platform image for platforms:", platformArg)
	err := doBuildTarget("multi", platformArg, config.Build.Options)
	if err != nil {
		tui.DisplayErrorMsg(fmt.Sprintf("Error building multi-platform image: %s", err.Error()))
		failOperation("Build failed")
	}
	tui.DisplaySuccessMsg("Multi-platform image built successfully")
}

func buildSinglePlatform(selectedName string) {
	settings := config.Build.Toolchains.Architectures[selectedName]
	cmdContext.Toolchain.Name = selectedName
	cmdContext.Toolchain.Platform = settings.Platform

	fmt.Printf("Building with toolchain '%s'...\n", selectedName)
	err := doBuildTarget(selectedName, settings.Platform, config.Build.Options)
	if err != nil {
		handleBuildError(selectedName, err)
	}
	tui.DisplaySuccessMsg(fmt.Sprintf("Application '%v' was built successfully with toolchain '%v'\n", config.App.Name, selectedName))
}

func handleBuildError(toolchainName string, err error) {
	tui.DisplayErrorMsg(fmt.Sprintf("Error building '%s': %s", toolchainName, err.Error()))
	failOperation("Build failed")
}

func doBuildTarget(architecture string, platform string, buildOptions configuration.BuildOptions) error {
	err := builder.BuildContainer(config.Build.Toolchains.Image, architecture, platform, projectRoot, config.App, buildOptions)

	if err != nil {
		return fmt.Errorf("invoking toolchain image '%v': %v", config.Build.Toolchains.Image, err)
	}

	return nil
}

func enableCrossCompilation(crossCompileSettings configuration.CrossCompileConfig) error {
	if err := builder.EnableMultiplatformBuild(crossCompileSettings); err != nil {
		return fmt.Errorf("setting up QEMU for cross-compilation failed: %w", err)
	}
	return nil
}

func validBuildArgsFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) == 0 {
		architectures := make([]string, 0, len(config.Build.Toolchains.Architectures))
		for k := range config.Build.Toolchains.Architectures {
			if strings.HasPrefix(k, toComplete) {
				architectures = append(architectures, k)
			}
		}
		return architectures, cobra.ShellCompDirectiveNoFileComp
	}

	return nil, cobra.ShellCompDirectiveNoFileComp
}
