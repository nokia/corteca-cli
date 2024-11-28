// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package cmd

import (
	"corteca/internal/configuration"
	"corteca/internal/configuration/templating"
	"corteca/internal/tui"
	"path/filepath"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

var regenCmd = &cobra.Command{
	Use:   "regen",
	Short: "Regenerate template files",
	Long:  "Regenerate template files",
	Run: func(cmd *cobra.Command, args []string) {
		requireProjectContext()
		doRegenTemplates(projectRoot)
	},
}

func init() {
	rootCmd.AddCommand(regenCmd)
}

func doRegenTemplates(destFolder string) {
	context := configuration.ToDictionary(config)
	fs := afero.NewOsFs()

	for srcFile, destFile := range config.Templates {
		if !filepath.IsAbs(destFile) {
			destFile = filepath.Join(destFolder, destFile)
		}
		assertOperation("regenerating template files", templating.RenderTemplateToFile(fs, srcFile, destFile, context))
		tui.LogNormal("%v was regenerated successfully", destFile)
	}
}
