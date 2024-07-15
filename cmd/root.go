// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package cmd

import (
	"corteca/internal/configuration"
	"corteca/internal/platform"
	"corteca/internal/templating"
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
	config           configuration.Settings
	configGlobal     configuration.Settings
	configSystem     configuration.Settings
	systemConfigRoot string
	userConfigRoot   string
	projectRoot      string
	distFolder       string
	configOverrides  []string
	languages        map[string]templating.TemplateInfo
	appVersion       string
)

var cmdContext struct {
	App            *configuration.AppSettings `yaml:"app,omitempty"`
	Arch           string                     `yaml:"arch,omitempty"`
	BuildArtifacts map[string]string          `yaml:"buildArtifacts,omitempty"`
	Deplenv        string                     `yaml:"deplenv,omitempty"`
	Device         struct {
		configuration.DeployDevice `yaml:",omitempty,inline"`
		Name                       string `yaml:"name,omitempty"`
	} `yaml:"device,omitempty"`
	Source struct {
		configuration.DownloadSource `yaml:",omitempty,inline"`
		Name                         string `yaml:"name,omitempty"`
	} `yaml:"source,omitempty"`
	Toolchain struct {
		Image    string `yaml:"image,omitempty"`
		Name     string `yaml:"name,omitempty"`
		Platform string `yaml:"platform,omitempty"`
	} `yaml:"toolchain,omitempty"`
	Build         *configuration.BuildSettings `yaml:"build,omitempty"`
	BuildArtifact string                       `yaml:"buildArtifact,omitempty"`
}

var (
	// mutually exclusive
	denvDev     bool
	denvStaging bool
	denvProd    bool
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
	rootCmd.PersistentFlags().StringVarP(&systemConfigRoot, "configRoot", "r", systemConfigRoot, "Override configuration root folder")
	rootCmd.PersistentFlags().StringVarP(&projectRoot, "projectRoot", "C", projectRoot, "Specify project root folder")
	// Deloyment stage
	rootCmd.PersistentFlags().BoolVar(&denvDev, "dev", false, "Deployment stage 'dev'")
	rootCmd.PersistentFlags().BoolVar(&denvStaging, "staging", false, "Deployment stage 'staging'")
	rootCmd.PersistentFlags().BoolVar(&denvProd, "prod", false, "Deployment stage 'prod'")
	rootCmd.MarkFlagsMutuallyExclusive("dev", "staging", "prod")
}

func Execute() {
	assertOperation("executing program", rootCmd.Execute())
}

func initConfiguration() {
	initEnvironment()
	config = configuration.NewConfiguration()

	// system global configuration
	assertOperation("reading system global configuration", config.ReadConfiguration(systemConfigRoot))
	configSystem = copystructure.Must(copystructure.Copy(config)).(configuration.Settings)

	// user global configuration
	if err := config.ReadConfiguration(userConfigRoot); !errors.Is(err, os.ErrNotExist) {
		assertOperation("reading user global configuration", err)
	}
	configGlobal = copystructure.Must(copystructure.Copy(config)).(configuration.Settings)

	if len(projectRoot) == 0 {
		localConfigDir, err := config.ReadConfigurationRecursive()
		assertOperation("reading current workdir configuration", err)
		projectRoot = localConfigDir
	} else {
		assertOperation("reading specified project root", config.ReadConfiguration(projectRoot))
	}
	// override config values
	assertOperation("parsing configuration overrides", overrideConfigValues())
	// TODO: validate configuration settings
}

func initEnvironment() {
	if denvStaging {
		cmdContext.Deplenv = "staging"
	} else if denvProd {
		cmdContext.Deplenv = "prod"
	} else {
		cmdContext.Deplenv = "dev" // default value
		denvDev = true
	}
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

func getLanguages() {
	if languages != nil {
		return
	}
	languages = make(map[string]templating.TemplateInfo)
	assertOperation("searching for system-wide language templates", templating.GetAvailableTemplates(languages, filepath.Join(systemConfigRoot, "templates")))
	assertOperation("searching for user language templates", templating.GetAvailableTemplates(languages, filepath.Join(userConfigRoot, "templates")))
}

func validateAppSettings(populateDefaults bool) error {
	getLanguages()
	var templ templating.TemplateInfo
	var found bool
	if config.App.Lang != "" {
		if templ, found = languages[config.App.Lang]; !found {
			return fmt.Errorf("no template for language '%v' was found", config.App.Lang)
		}
	} else {
		return fmt.Errorf("no programming language has been specified")
	}
	if config.App.Title == "" {
		return fmt.Errorf("no application title has been specified")
	}
	if config.App.Name == "" {
		return fmt.Errorf("no application name has been specified")
	}
	if config.App.Version == "" {
		return fmt.Errorf("no application version has been specified")
	}
	if config.App.FQDN == "" {
		return fmt.Errorf("no application FQDN has been specified")
	} else if config.App.DUID == "" {
		config.App.DUID = generateDUID(config.App.FQDN)
		fmt.Printf("Generated application DUID: %v\n", config.App.DUID)
	}
	// if we have found a template, we can proceed validating app options; otherwise return with no errors
	if !found {
		return nil
	}
	for _, option := range templ.Options {
		if _, found = config.App.Options[option.Name]; found {
			value, err := validateOptionValue(config.App.Options[option.Name], option)
			if err != nil {
				return err
			}
			config.App.Options[option.Name] = value
		} else if populateDefaults {
			config.App.Options[option.Name] = option.Default
		}
	}
	// populate app dependencies
	if populateDefaults {
		config.App.Dependencies.Compile = append(config.App.Dependencies.Compile, templ.Dependencies.Compile...)
		config.App.Dependencies.Runtime = append(config.App.Dependencies.Runtime, templ.Dependencies.Runtime...)
	}
	return nil
}

func assertOperation(operation string, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while %v: %v\n", operation, err.Error())
		os.Exit(1)
	}
}

func failOperation(msg string) {
	fmt.Fprintf(os.Stderr, "Fatal error: %s\n", msg)
	os.Exit(1)
}

func requireProjectContext() {
	if projectRoot == "" {
		failOperation("must be run inside a project context")
		os.Exit(1)
	}
	cmdContext.App = &config.App
	cmdContext.Build = &config.Build
}

func requireBuildArtifact() {
	requireProjectContext()
	distFolder = filepath.Join(projectRoot, distFolderName)

	cmdContext.BuildArtifacts = make(map[string]string)

	pattern := filepath.Join(distFolder, fmt.Sprintf("%v*.tar.gz", config.App.Name))
	distFiles, _ := filepath.Glob(pattern)

	// Compile a regular expression to extract the CPU architecture from the filename.
	cpuArchRegex := regexp.MustCompile(fmt.Sprintf(`^%s-(?:[^-]+)-([^-.]+)-rootfs\.tar\.gz$`, regexp.QuoteMeta(config.App.Name)))

	// Iterate over each file found in the distribution folder that matches the pattern.
	for _, distFile := range distFiles {
		filename := filepath.Base(distFile)
		matches := cpuArchRegex.FindStringSubmatch(filename)

		// If the filename contains a CPU architecture, process it.
		if len(matches) > 1 {
			cpuArch := matches[1]

			// If we've already selected a build artifact for this CPU architecture, compare modification times.
			if curArtifactName, ok := cmdContext.BuildArtifacts[cpuArch]; ok {
				curArtifactInfo, err := os.Stat(curArtifactName)
				if err != nil {
					failOperation(fmt.Sprintf("stating artifact %s failed: %v", curArtifactName, err))
				}

				distFileInfo, err := os.Stat(distFile)
				if err != nil {
					failOperation(fmt.Sprintf("stating artifact %s failed: %v", distFile, err))
				}

				// Update the selection if the new candidate is more recent.
				if distFileInfo.ModTime().After(curArtifactInfo.ModTime()) {
					cmdContext.BuildArtifacts[cpuArch] = distFile
				}
			} else {
				// Found this CPU architecture for the first time, assign current artifact.
				cmdContext.BuildArtifacts[cpuArch] = distFile
			}
		}
	}

	if len(cmdContext.BuildArtifacts) == 0 {
		failOperation("no build artifacts found")
	}
}

func generateDUID(FQDN string) string {
	return uuid.NewSHA1(uuid.NameSpaceURL, []byte(FQDN)).String()
}
