// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package cmd

import (
	"github.com/nokia/corteca-cli/internal/configuration"
	specs "github.com/nokia/corteca-cli/internal/configuration/runtimeSpec"
	"github.com/nokia/corteca-cli/internal/platform"
	"github.com/nokia/corteca-cli/internal/tui"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/mitchellh/copystructure"
	"github.com/spf13/cobra"
)

const distFolderName = "dist"

var (
	cpuShares      = uint64(1024)
	cpuQuota       = int64(5)
	cpuPeriod      = uint64(100)
	memLimit       = int64(15728640)
	memReservation = int64(15728640)
	memSwap        = int64(31457280)
)

var (
	config           configuration.Settings
	configGlobal     configuration.Settings
	configSystem     configuration.Settings
	systemConfigRoot string
	userConfigRoot   string
	projectRoot      string
	distFolder       string
	artifact         string
	configOverrides  []string
	templates        map[string]configuration.TemplateInfo
	appVersion       string
	skipLocalConfig  bool
	noRegen          bool
)

var rootCmd = &cobra.Command{
	Use:              "corteca",
	Short:            "Nokia Corteca Developer Toolkit",
	Long:             `The Corteca Developer Toolkit facilitates bootstrapping, building and deploying containerized applications for Nokia BroadBand Devices`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) { initConfiguration() },
	Version:          appVersion,
}

func init() {
	systemConfigRoot, userConfigRoot = platform.SetConfigPaths()
	rootCmd.SetVersionTemplate("{{.Short}} v{{.Version}}\n")
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	rootCmd.PersistentFlags().StringArrayVarP(&configOverrides, "config", "c", []string{}, "Override a configuration value in the form of a 'key=value' pair")
	_ = rootCmd.RegisterFlagCompletionFunc("config", validConfigArgsFunc)
	rootCmd.PersistentFlags().StringVarP(&systemConfigRoot, "configRoot", "r", systemConfigRoot, "Override configuration root folder")
	_ = rootCmd.RegisterFlagCompletionFunc("configRoot", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveFilterDirs
	})
	rootCmd.PersistentFlags().StringVarP(&projectRoot, "projectRoot", "C", projectRoot, "Specify project root folder")
	_ = rootCmd.RegisterFlagCompletionFunc("projectRoot", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveFilterDirs
	})
	rootCmd.PersistentFlags().BoolVar(&tui.DisableColoredOutput, "no-color", false, "Disables colored stdout|stderr output")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func readLocalConfiguration() {
	if len(projectRoot) == 0 {
		localConfigDir, err := config.ReadConfigurationRecursive()
		assertOperation("reading current workdir configuration", err)
		projectRoot = localConfigDir
	} else {
		assertOperation("reading specified project root", config.ReadConfiguration(projectRoot))
	}
}

func initConfiguration() {
	tui.DefineOutputColor()
	config = configuration.NewConfiguration()

	// system global configuration
	assertOperation("reading system global configuration", config.ReadConfiguration(systemConfigRoot))
	configSystem = copystructure.Must(copystructure.Copy(config)).(configuration.Settings)

	// user global configuration
	if err := config.ReadConfiguration(userConfigRoot); !errors.Is(err, os.ErrNotExist) {
		assertOperation("reading user global configuration", err)
	}
	configGlobal = copystructure.Must(copystructure.Copy(config)).(configuration.Settings)

	if !skipLocalConfig {
		readLocalConfiguration()
	}
	// override config values
	assertOperation("parsing configuration overrides", overrideConfigValues())
	// TODO: validate configuration settings
}

func overrideConfigValues() error {
	for _, entry := range configOverrides {
		key, val, found := strings.Cut(entry, "=")
		if !found {
			return fmt.Errorf("invalid syntax for option '%v'; use 'key=value'", entry)
		}
		if err := config.WriteField(key, val, false); err != nil {
			return err
		}
	}
	return nil
}

func getTemplates() {
	if templates != nil {
		return
	}
	templates = make(map[string]configuration.TemplateInfo)
	assertOperation("searching for system-wide language templates", configuration.GetAvailableTemplates(templates, filepath.Join(systemConfigRoot, "templates")))
	assertOperation("searching for user language templates", configuration.GetAvailableTemplates(templates, filepath.Join(userConfigRoot, "templates")))
}

func validateAppSettings() error {
	if config.App.Name == "" {
		return fmt.Errorf("no application name has been specified")
	}
	if config.App.Version == "" {
		return fmt.Errorf("no application version has been specified")
	}
	if config.App.DUID == "" {
		return fmt.Errorf("DUID has not been generated successfully")
	}
	if len(config.App.Entrypoint) == 0 {
		config.App.Entrypoint = append(config.App.Entrypoint, filepath.ToSlash(filepath.Join("/bin", config.App.Name)))
	}
	return nil
}

func assertOperation(operation string, err error) {
	if err != nil {
		tui.LogError("Error while %v: %v", operation, err.Error())
		os.Exit(1)
	}
}

func failOperation(msg string) {
	tui.LogError("Fatal error: %s", msg)
	os.Exit(1)
}

func defaultRuntimeSpec(name string) specs.Spec {

	return specs.Spec{
		Hooks: &specs.Hooks{
			Prestart: []specs.Hook{
				{Path: "/bin/prepare_container.sh"},
			},
			Poststop: []specs.Hook{
				{Path: "/bin/cleanup_container.sh"},
			},
		},
		Hostname: name,
		Linux: &specs.Linux{
			Resources: &specs.LinuxResources{
				CPU: &specs.LinuxCPU{
					Shares: &cpuShares,
					Quota:  &cpuQuota,
					Period: &cpuPeriod,
				},
				Memory: &specs.LinuxMemory{
					Limit:       &memLimit,
					Reservation: &memReservation,
					Swap:        &memSwap,
				},
			},
		},
		Mounts: []specs.Mount{
			{
				Destination: "/opt",
				Type:        "bind",
				Source:      "/var/run/ubus-session",
				Options:     []string{"rbind", "rw"},
			},
		},
		Version: "1.0.0",
	}
}

func requireProjectContext() {
	if projectRoot == "" {
		failOperation("must be run inside a project context")
		os.Exit(1)
	}
	configuration.GetCmdContext().App = &config.App
	configuration.GetCmdContext().Build = &config.Build
}


func requireBuildArtifact() {
	if artifact != "" {
		if _, err := os.Stat(artifact); errors.Is(err, os.ErrNotExist) {
			failOperation(fmt.Sprintf("file %s not found", artifact))
		}
		distFolder = filepath.Dir(artifact)
	} else {
		requireProjectContext()
		distFolder = filepath.Join(projectRoot, distFolderName)
		var buildArtifacts []string
		patterns := []string{"*.tar.gz", "*.tar", "*.zip"}
		for _, pattern := range patterns {
			files, _ := filepath.Glob(filepath.Join(distFolder, pattern))
			buildArtifacts = append(buildArtifacts, files...)
		}

		if len(buildArtifacts) == 0 {
			failOperation("no build artifacts found")
		} else if len(buildArtifacts) > 1 {
			var err error
			slices.Sort(buildArtifacts)
			artifact, err = tui.PromptForSelection("Select artifact to publish", buildArtifacts, buildArtifacts[0])
			if err != nil {
				failOperation("artifact selection cancelled")
			}
		} else {
			artifact = buildArtifacts[0]
		}
	}
	configuration.GetCmdContext().Artifact = artifact
}

func generateDUID(input string) string {
	if input == "" {
		return ""
	}
	return uuid.NewSHA1(uuid.NameSpaceURL, []byte(input)).String()
}
