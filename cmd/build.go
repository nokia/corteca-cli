// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package cmd

import (
	"corteca/internal/builder"
	"corteca/internal/configuration"
	"corteca/internal/fsutil"
	"corteca/internal/packager"
	"corteca/internal/tui"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	rootfsPath string
	err        error
)

const (
	rootfsTarballName = "rootfs.tar.gz"
)

var buildCmd = &cobra.Command{
	Use:   "build [ARCHITECTURE]",
	Short: "Build an application",
	Long:  `Build the application using the appropriate build architecture`,
	Example: `#Build the application for specific architeture (armv7l) as OCI image
corteca build armv7l --config build.options.outputType=oci`,
	Args:              cobra.MaximumNArgs(1),
	ValidArgsFunction: validBuildArgsFunc,
	Run: func(cmd *cobra.Command, args []string) {
		architecture := ""
		if len(args) > 0 {
			architecture = args[0]
		}

		doBuildApp(architecture)
	},
}

func init() {
	buildCmd.PersistentFlags().BoolVar(&noRegen, "no-regen", false, "Skip regeneration of templates")
	buildCmd.PersistentFlags().StringVarP(&rootfsPath, "rootfs", "", "", "Specify prebuilt root filesystem")
	rootCmd.AddCommand(buildCmd)
}

func generateBuildInfo(version, appVersion string) map[string]string {
	return map[string]string{
		"com.nokia.corteca.version":     version,
		"com.nokia.corteca.app.version": appVersion,
		"com.nokia.corteca.app.created": time.Now().UTC().Format(time.RFC3339),
	}
}

func doBuildApp(selectedArchitecture string) {
	requireProjectContext()
	if rootfsPath != "" {
		info, err := os.Stat(rootfsPath)
		assertOperation("reading prebuild root fs path", err)
		if info.IsDir() {
			failOperation("Rootfs from folder is not supported")
		}
	}
	if !noRegen {
		doRegenTemplates(projectRoot, selectedArchitecture)
	}

	selectedArchitecture, settings := validateArchitecture(selectedArchitecture)
	outputType := strings.ToLower(config.Build.Options.OutputType)
	configuration.CmdContext.Arch = selectedArchitecture
	configuration.CmdContext.Platform = settings.Platform
	buildMetadata := generateBuildInfo(appVersion, config.App.Version)

	// STEP 0: create necessary folders
	tmpBuildPath, err := os.MkdirTemp("", "corteca_build-")
	assertOperation("creating temporary folder", err)

	rootfsBuildPath := filepath.Join(tmpBuildPath, "rootfs")
	err = os.MkdirAll(rootfsBuildPath, 0777)
	assertOperation("creating rootfs folder", err)

	distPath := filepath.Join(projectRoot, "dist")
	assertOperation("creating dist directory", os.MkdirAll(distPath, 0755))

	// STEP 1: build rootfs
	if rootfsPath == "" {
		assertOperation(fmt.Sprintf("building '%s'", selectedArchitecture), builder.BuildRootFS(projectRoot, rootfsBuildPath, settings, config.App, config.Build))
	} else {
		fsutil.ExtractTarball(rootfsPath, rootfsBuildPath)
	}

	// STEP 2: validate rootfs
	// TODO: skip if a flag to skip validation has been provided
	assertOperation("validating rootfs", packager.ValidateRootFS(rootfsBuildPath, selectedArchitecture, config.App))

	// STEP 3: annotate rootfs with build information
	assertOperation("adding annotations", packager.AnnotateRootFS(rootfsBuildPath, config.App, buildMetadata))

	// STEP 4: commpress rootfs into a tarball
	tmprootfsTarGzPath := filepath.Join(tmpBuildPath, rootfsTarballName)
	assertOperation("compressing rootfs", packager.CompressRootfs(rootfsBuildPath, tmprootfsTarGzPath))

	// STEP 5: create amd commpress runtime config into a tarball
	assertOperation("create and compress runtime config", packager.CreateAndCompressRuntimeConfig(tmpBuildPath, config.App, buildMetadata))

	// STEP 6: package rootfs depending on the output type
	switch outputType {
	case "oci":
		assertOperation("packaging OCI", packager.PackageOCI(tmpBuildPath, distPath, selectedArchitecture, settings.Platform, tmprootfsTarGzPath, config.App))
	case "rootfs":
		assertOperation("packaging RootFS", packager.PackageRootFS(tmpBuildPath, projectRoot, distPath, selectedArchitecture, outputType, config.App))
	default:
		failOperation(fmt.Sprintf("Unsupported output type: %q", outputType))
	}

	tui.DisplaySuccessMsg(fmt.Sprintf("Application '%v' was built successfully for '%v'", config.App.Name, selectedArchitecture))
}

func validBuildArgsFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) == 0 {
		architectures := make([]string, 0, len(config.Build.Architectures))
		for k := range config.Build.Architectures {
			if strings.HasPrefix(k, toComplete) {
				architectures = append(architectures, k)
			}
		}
		return architectures, cobra.ShellCompDirectiveNoFileComp
	}

	return nil, cobra.ShellCompDirectiveNoFileComp
}

func validateArchitecture(selectedArchitecture string) (string, configuration.ArchitectureSettings) {
	if selectedArchitecture == "" {
		return config.Build.Default, config.Build.Architectures[config.Build.Default]
	} else if _, ok := config.Build.Architectures[selectedArchitecture]; !ok {
		failOperation(fmt.Sprintf("no configuration present for toolchain: \"%v\"", selectedArchitecture))
	}
	return selectedArchitecture, config.Build.Architectures[selectedArchitecture]
}
