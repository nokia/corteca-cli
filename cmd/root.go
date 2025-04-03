// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package cmd

import (
	"corteca/internal/configuration"
	specs "corteca/internal/configuration/runtimeSpec"
	"corteca/internal/platform"
	"corteca/internal/tui"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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
	config            configuration.Settings
	configGlobal      configuration.Settings
	configSystem      configuration.Settings
	systemConfigRoot  string
	userConfigRoot    string
	projectRoot       string
	distFolder        string
	specifiedArtifact string
	configOverrides   []string
	templates         map[string]configuration.TemplateInfo
	appVersion        string
	skipLocalConfig   bool
	noRegen           bool
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
	rootCmd.RegisterFlagCompletionFunc("config", validConfigArgsFunc)
	rootCmd.PersistentFlags().StringVarP(&systemConfigRoot, "configRoot", "r", systemConfigRoot, "Override configuration root folder")
	rootCmd.RegisterFlagCompletionFunc("configRoot", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveFilterDirs
	})
	rootCmd.PersistentFlags().StringVarP(&projectRoot, "projectRoot", "C", projectRoot, "Specify project root folder")
	rootCmd.RegisterFlagCompletionFunc("projectRoot", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveFilterDirs
	})
	rootCmd.PersistentFlags().BoolVar(&tui.DisableColoredOutput, "no-color", false, "Disables colored stdout|stderr output")
}

func Execute() {
	rootCmd.Execute()
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
	if config.App.Entrypoint == "" {
		config.App.Entrypoint = filepath.Join("/bin", config.App.Name)
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
	configuration.CmdContext.App = &config.App
	configuration.CmdContext.Build = &config.Build
}

func splitSpecifiedArtifact(specifiedArtifact string) (arch, imgType, path string) {
	artifactInfo := strings.SplitN(specifiedArtifact, ":", 3)
	if len(artifactInfo) < 3 || artifactInfo[2] == "" {
		failOperation("architecture, image type or path to artifact is missing")
	}
	if !(filepath.Ext(artifactInfo[2]) == ".gz" || filepath.Ext(artifactInfo[2]) == ".tar") {
		failOperation("artifact file should be of type \".tar.gz\" or \".tar\"")
	}
	return strings.ToLower(artifactInfo[0]), strings.ToLower(artifactInfo[1]), artifactInfo[2]
}

func getAppNameFromArtifact(artifactPath string) string {
	artifactName := filepath.Base(artifactPath)
	splitedArtifactName := strings.SplitN(artifactName, "-", 2)
	// If artifact name doesn't contain hyphens we consider the form "appName.tar.gz", else we consider App.Name the string up to the first hyphen.
	if len(splitedArtifactName) == 1 {
		return strings.TrimSuffix(splitedArtifactName[0], ".tar.gz")
	} else {
		return splitedArtifactName[0]
	}
}

func requireBuildArtifact() {
	configuration.CmdContext.BuildArtifacts = make(map[string]string)
	if specifiedArtifact != "" {
		artifactArch, artifactType, artifactPath := splitSpecifiedArtifact(specifiedArtifact)
		if _, err := os.Stat(artifactPath); errors.Is(err, os.ErrNotExist) {
			failOperation(fmt.Sprintf("file %s not found", artifactPath))
		}
		configuration.CmdContext.BuildArtifacts[artifactArch+"-"+artifactType] = artifactPath
		distFolder = filepath.Dir(artifactPath)
		configuration.CmdContext.Arch = artifactArch
		// Set necessary build field for deployment
		configuration.CmdContext.Build = &config.Build
		configuration.CmdContext.Build.Options.OutputType = artifactType
		// Set necessary app fields for deployment
		configuration.CmdContext.App = &config.App

		if skipLocalConfig || len(configuration.CmdContext.App.DUID) == 0 {
			configuration.CmdContext.App.DUID = generateDUID(artifactPath)
		}
		if skipLocalConfig || len(configuration.CmdContext.App.Name) == 0 {
			configuration.CmdContext.App.Name = getAppNameFromArtifact(artifactPath)
		}

		return
	}
	requireProjectContext()

	distFolder = filepath.Join(projectRoot, distFolderName)

	rootfsPattern := filepath.Join(distFolder, fmt.Sprintf("%v*-rootfs.tar.gz", config.App.Name))
	ociPattern := filepath.Join(distFolder, fmt.Sprintf("%v*-oci.tar", config.App.Name))

	rootfsFiles, _ := filepath.Glob(rootfsPattern)
	ociFiles, _ := filepath.Glob(ociPattern)

	// Compile a common regular expression to extract the CPU architecture from the filename.
	commonArchRegex := regexp.MustCompile(fmt.Sprintf(`^%s-(?:[^-]+)-([^-.]+)-(rootfs|oci)\.(tar\.gz|tar)$`, regexp.QuoteMeta(config.App.Name)))
	matchArchitectures(commonArchRegex, rootfsFiles, "rootfs")
	matchArchitectures(commonArchRegex, ociFiles, "oci")

	if len(configuration.CmdContext.BuildArtifacts) == 0 {
		failOperation("no build artifacts found")
	}
}

func matchArchitectures(archRegex *regexp.Regexp, distFiles []string, artifactType string) {
	for _, distFile := range distFiles {
		filename := filepath.Base(distFile)
		matches := archRegex.FindStringSubmatch(filename)

		// If the filename contains a CPU architecture, process it.
		if len(matches) < 2 {
			continue
		}
		cpuArch := matches[1]
		if curArtifactName, ok := configuration.CmdContext.BuildArtifacts[cpuArch+"-"+artifactType]; ok {
			curArtifactInfo, err := os.Stat(curArtifactName)
			if err != nil {
				failOperation(fmt.Sprintf("stating artifact %s failed: %v", curArtifactName, err))
			}

			distFileInfo, err := os.Stat(distFile)
			if err != nil {
				failOperation(fmt.Sprintf("stating artifact %s failed: %v", distFile, err))
			}

			// Update the selection if the new candidate is more recent and continue the loop
			if distFileInfo.ModTime().After(curArtifactInfo.ModTime()) {
				configuration.CmdContext.BuildArtifacts[cpuArch+"-"+artifactType] = distFile
			}

			continue
		}
		configuration.CmdContext.BuildArtifacts[cpuArch+"-"+artifactType] = distFile
	}
}

func generateDUID(input string) string {
	if input == "" {
		return ""
	}
	return uuid.NewSHA1(uuid.NameSpaceURL, []byte(input)).String()
}
