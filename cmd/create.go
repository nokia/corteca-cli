// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package cmd

import (
	"corteca/internal/configuration"
	"corteca/internal/tui"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

var (
	skipPrompts bool
)

var createCmd = &cobra.Command{
	Use:   "create DESTFOLDER",
	Short: "Create a new application skeleton",
	Long:  `Create a new application skeleton for the specified target programming language. If no destination folder is provided, current one is assumed`,
	Args:  cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveFilterDirs
	},
	Run: func(cmd *cobra.Command, args []string) { doCreateApp(args[0]) },
}

func init() {
	createCmd.PersistentFlags().BoolVar(&skipPrompts, "skipPrompts", false, "Skip user-prompts for settings not provided via command line")
	rootCmd.AddCommand(createCmd)
}

func doCreateApp(destFolder string) {
	if !skipPrompts {
		// Prompt for application config values that where not given:
		collectAppSettings()
	}
	assertOperation("validating application settings", validateAppSettings(true))

	// determine destination folder
	assertOperation("creating destination folder", os.MkdirAll(destFolder, 0777))

	templ := languages[config.App.Lang]
	// convert to map[string]any to use json-tag fields
	context := configuration.ToDictionary(config)
	fs := afero.NewOsFs()

	tui.LogNormal("Using template from: %v", languages[config.App.Lang].Path)
	assertOperation("rendering language template", configuration.GenerateTemplate(fs, templ, destFolder, context))
	assertOperation("writing local config", config.WriteConfiguration(destFolder, &configGlobal))
	tui.LogNormal("Generated a new %v application in '%v'", config.App.Lang, destFolder)
}

// Helpers

func collectAppSettings() {
	getLanguages()
	langNames := []string{}
	for lang := range languages {
		langNames = append(langNames, lang)
	}

	// app language
	var err error
	for config.App.Lang == "" {
		tui.DisplayHelpMsg("Choose application programming language; each language may require different options")
		config.App.Lang, err = tui.PromptForSelection("*Mandatory* Language", langNames, "")
		assertOperation("choosing language", err)
	}
	// title
	if config.App.Title == "" {
		tui.DisplayHelpMsg("Specify application title; can be comprised by multiple words")
		config.App.Title, err = tui.PromptForValue("*Mandatory* Application title", "")
		assertOperation("specifying application title", err)
	}
	// name
	if config.App.Name == "" {
		tui.DisplayHelpMsg("Specify application name; a single word identifier that cannot contain spaces")
		defaultName := strings.Replace(strings.ToLower(config.App.Title), " ", "_", -1)
		config.App.Name, err = tui.PromptForValue("Application name", defaultName)
		assertOperation("specifying application name", err)
		if strings.Contains(config.App.Name, " ") {
			failOperation("application name cannot contain spaces")
		}
	}
	//FQDN
	if config.App.FQDN == "" {
		tui.DisplayHelpMsg("Specify application Fully Qualified Domain Name; should be in the form of \"domain.example.com\"")
		config.App.FQDN, err = tui.PromptForValue("*Mandatory* FQDN", "")
		assertOperation("specifying FQDN", err)
	}
	// author
	if config.App.Author == "" {
		tui.DisplayHelpMsg("Specify application author")
		config.App.Author, err = tui.PromptForValue("Author", "")
		assertOperation("specifying author", err)
	}
	// description
	if config.App.Description == "" {
		tui.DisplayHelpMsg("Specify application full description")
		config.App.Description, err = tui.PromptForValue("Description", "")
		assertOperation("specifying description", err)
	}
	// version
	if config.App.Version == "" {
		tui.DisplayHelpMsg("Specify application version; should be in the form of X.X.X")
		config.App.Version, err = tui.PromptForValue("*Mandatory* Version", "")
		assertOperation("specifying version", err)
	}
	// custom language options
	options := languages[config.App.Lang].Options
	for _, option := range options {
		if _, exists := config.App.Options[option.Name]; !exists {
			config.App.Options[option.Name], err = promptForOption(option)
			assertOperation("selecting custom language option", err)
		}
	}
}

func validateOptionValue(value any, option configuration.TemplateCustomOption) (any, error) {
	optionType := strings.ToLower(option.Type)
	valueType := reflect.TypeOf(value)
	if optionType == configuration.BoolOption {
		if valueType.Kind() == reflect.String {
			return strconv.ParseBool(value.(string))
		} else if valueType.Kind() != reflect.Bool {
			return nil, fmt.Errorf("expected bool, got %v", valueType.Name())
		}
	} else if optionType == configuration.TextOption || optionType == configuration.ChoiceOption {
		if valueType.Kind() != reflect.String {
			return fmt.Sprintf("%v", value), nil
		}
	}
	return value, nil
}

func promptForOption(option configuration.TemplateCustomOption) (any, error) {
	optionType := strings.ToLower(option.Type)
	if optionType == configuration.BoolOption {
		return tui.PromptForConfirm(option.Description, option.Default.(bool))
	} else if optionType == configuration.TextOption {
		return tui.PromptForValue(option.Description, option.Default.(string))
	} else if optionType == configuration.ChoiceOption {
		return tui.PromptForSelection(option.Description, option.Values, option.Default.(string))
	}
	return nil, fmt.Errorf("unknown option type '%v'", optionType)
}
