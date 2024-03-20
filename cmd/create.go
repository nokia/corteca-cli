// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package cmd

import (
	"corteca/internal/configuration"
	"corteca/internal/templating"
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
	Run:   func(cmd *cobra.Command, args []string) { doCreateApp(args[0]) },
}

func init() {
	createCmd.PersistentFlags().BoolVar(&skipPrompts, "skipPrompts", false, "Skip user-prompts for settings not provided via command line")
	rootCmd.AddCommand(createCmd)
}

func doCreateApp(destFolder string) {
	if !skipPrompts {
		// Prompt for application config values that where not given:
		assertOperation("collecting application settings", collectAppSettings())
	}
	assertOperation("validating application settings", validateAppSettings(true))

	// determine destination folder
	assertOperation("creating destination folder", os.MkdirAll(destFolder, 0777))

	templ := languages[config.App.Lang]
	// convert to map[string]any to use json-tag fields
	context := configuration.ToDictionary(config)
	fs := afero.NewOsFs()

	fmt.Fprintf(os.Stdout, "Using template from: %v\n", languages[config.App.Lang].Path)
	assertOperation("rendering language template", templating.GenerateTemplate(fs, templ, destFolder, context))
	assertOperation("writing local config", config.WriteConfiguration(destFolder, &configGlobal))
	fmt.Fprintf(os.Stdout, "Generated a new %v application in '%v'\n", config.App.Lang, destFolder)
}

// Helpers

func collectAppSettings() error {
	getLanguages()
	langNames := []string{}
	for lang := range languages {
		langNames = append(langNames, lang)
	}

	// app language
	var err error
	for config.App.Lang == "" {
		tui.DisplayHelpMsg("Choose application programming language; each language may require different options")
		config.App.Lang, err = tui.PromptForSelection("Language", langNames, "")
		if err != nil {
			return err
		}
	}
	// title
	if config.App.Title == "" {
		tui.DisplayHelpMsg("Specify application title; can be comprised by multiple words")
		config.App.Title, err = tui.PromptForValue("Application title", "")
		if err != nil {
			return err
		}
	}
	// name
	if config.App.Name == "" {
		tui.DisplayHelpMsg("Specify application name; a single word identifier that cannot contain spaces")
		defaultName := strings.Replace(strings.ToLower(config.App.Title), " ", "_", -1)
		config.App.Name, err = tui.PromptForValue("Application name", defaultName)
		if err != nil {
			return err
		}
		if strings.Contains(config.App.Name, " ") {
			return fmt.Errorf("application name cannot contain spaces")
		}
	}
	//FQDN
	if config.App.FQDN == "" {
		tui.DisplayHelpMsg("Specify application Fully Qualified Domain Name; should be in the form of \"domain.example.com\"")
		config.App.FQDN, err = tui.PromptForValue("FQDN", "")
		if err != nil {
			return err
		}
	}
	// author
	if config.App.Author == "" {
		tui.DisplayHelpMsg("Specify application author")
		config.App.Author, err = tui.PromptForValue("Author", "")
		if err != nil {
			return err
		}
	}
	// description
	if config.App.Description == "" {
		tui.DisplayHelpMsg("Specify application full description")
		config.App.Description, err = tui.PromptForValue("Description", "")
		if err != nil {
			return err
		}
	}
	// version
	if config.App.Version == "" {
		tui.DisplayHelpMsg("Specify application version; should be in the form of X.X.X")
		config.App.Version, err = tui.PromptForValue("Version", "")
		if err != nil {
			return err
		}
	}
	// custom language options
	options := languages[config.App.Lang].Options
	for _, option := range options {
		if _, exists := config.App.Options[option.Name]; !exists {
			config.App.Options[option.Name], err = promptForOption(option)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func validateOptionValue(value any, option templating.TemplateCustomOption) (any, error) {
	optionType := strings.ToLower(option.Type)
	valueType := reflect.TypeOf(value)
	if optionType == templating.BoolOption {
		if valueType.Kind() == reflect.String {
			return strconv.ParseBool(value.(string))
		} else if valueType.Kind() != reflect.Bool {
			return nil, fmt.Errorf("expected bool, got %v", valueType.Name())
		}
	} else if optionType == templating.TextOption || optionType == templating.ChoiceOption {
		if valueType.Kind() != reflect.String {
			return fmt.Sprintf("%v", value), nil
		}
	}
	return value, nil
}

func promptForOption(option templating.TemplateCustomOption) (any, error) {
	optionType := strings.ToLower(option.Type)
	if optionType == templating.BoolOption {
		return tui.PromptForConfirm(option.Description, option.Default.(bool))
	} else if optionType == templating.TextOption {
		return tui.PromptForValue(option.Description, option.Default.(string))
	} else if optionType == templating.ChoiceOption {
		return tui.PromptForSelection(option.Description, option.Values, option.Default.(string))
	}
	return nil, fmt.Errorf("unknown option type '%v'", optionType)
}
