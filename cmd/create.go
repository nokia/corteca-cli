// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package cmd

import (
	"corteca/internal/configuration"
	"corteca/internal/tui"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

var (
	skipPrompts bool
	lang        string
	fqdn        string
)

var createCmd = &cobra.Command{
	Use:     "create DESTFOLDER",
	Short:   "Create a new application skeleton",
	Long:    "Create a new application skeleton for the specified target programming language\nIf no destination folder is provided, current one is assumed",
	Example: "",
	Args:    cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveFilterDirs
	},
	Run: func(cmd *cobra.Command, args []string) { doCreateApp(args[0]) },
}

func init() {
	createCmd.PersistentFlags().BoolVar(&skipPrompts, "skipPrompts", false, "Skip user-prompts for settings not provided via command line")
	createCmd.PersistentFlags().StringVarP(&lang, "lang", "l", "", "Specify the language to use")
	createCmd.PersistentFlags().StringVarP(&fqdn, "fqdn", "f", "", "Specify the FQDN")
	rootCmd.AddCommand(createCmd)
}

func doCreateApp(destFolder string) {
	if !skipPrompts {
		// Prompt for application config values that where not given:
		collectAppSettings()
	}
	assertOperation("populating application dependencies", populateAppDeps())
	config.App.Runtime = defaultRuntimeSpec(config.App.Name)
	config.App.DUID = generateDUID(fqdn)
	assertOperation("validating application settings", validateAppSettings())

	// determine destination folder
	assertOperation("creating destination folder", os.MkdirAll(destFolder, 0777))

	langTemplate := templates[lang]
	// convert to map[string]any to use json-tag fields
	context := configuration.ToDictionary(config)
	fs := afero.NewOsFs()

	tui.LogNormal("Using template from: %v", templates[lang].Path)
	assertOperation("rendering language template", configuration.GenerateTemplate(fs, langTemplate, destFolder, context, templates, nil, config.Templates))
	assertOperation("writing local config", config.WriteConfiguration(destFolder, &configGlobal))
	tui.LogNormal("Generated a new %v application in '%v'", lang, destFolder)
}

// Helpers

func populateAppDeps() error {
	getTemplates()
	var templ configuration.TemplateInfo
	var found bool
	if templ, found = templates[lang]; !found {
		return fmt.Errorf("no template for language '%v' was found", lang)
	}
	config.App.Dependencies.Compile = append(config.App.Dependencies.Compile, templ.Dependencies.Compile...)
	config.App.Dependencies.Runtime = append(config.App.Dependencies.Runtime, templ.Dependencies.Runtime...)

	return nil
}

func collectAppSettings() {
	getTemplates()
	langNames := []string{}
	for lang := range templates {
		if !strings.HasPrefix(lang, "_") {
			langNames = append(langNames, lang)
		}
	}

	// app language
	var err error
	for lang == "" {
		tui.DisplayHelpMsg("Choose application programming language; each language may require different options")
		lang, err = tui.PromptForSelection("*Mandatory* Language", langNames, "")
		assertOperation("choosing language", err)
	}
	// name
	if config.App.Name == "" {
		tui.DisplayHelpMsg("Specify application name; a single word identifier that cannot contain spaces")
		config.App.Name, err = tui.PromptForValue("Application name", "")
		assertOperation("specifying application name", err)
		if strings.Contains(config.App.Name, " ") {
			failOperation("application name cannot contain spaces")
		}
	}
	//FQDN
	if fqdn == "" {
		tui.DisplayHelpMsg("Specify application Fully Qualified Domain Name; should be in the form of \"domain.example.com\"")
		fqdn, err = tui.PromptForValue("*Mandatory* FQDN", "")
		assertOperation("specifying FQDN", err)
	}
	// author
	if config.App.Author == "" {
		tui.DisplayHelpMsg("Specify application author")
		config.App.Author, err = tui.PromptForValue("Author", "")
		assertOperation("specifying author", err)
	}
	// version
	if config.App.Version == "" {
		tui.DisplayHelpMsg("Specify application version; should be in the form of X.X.X")
		config.App.Version, err = tui.PromptForValue("*Mandatory* Version", "")
		assertOperation("specifying version", err)
	}
}
