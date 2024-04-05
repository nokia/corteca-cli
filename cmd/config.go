// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package cmd

import (
	"corteca/internal/configuration"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	affectGlobalConfig bool
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Read/write configuration values",
	Long:  `Read/write configuration values`,
	Args:  cobra.MaximumNArgs(1),
}

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Long:  "Show all configuration settings in json format",
	Run:   func(cmd *cobra.Command, args []string) { doShowConfig() },
}

var getCmd = &cobra.Command{
	Use:   "get key",
	Short: "Read a configuration value",
	Long:  "Read a configuration value",
	Args:  cobra.ExactArgs(1),
	Run:   func(cmd *cobra.Command, args []string) { doGetConfigValue(args[0]) },
}

var setCmd = &cobra.Command{
	Use:   "set key value",
	Short: "Set a configuration value",
	Long:  "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	Run:   func(cmd *cobra.Command, args []string) { doSetConfigValue(args[0], args[1]) },
}

func init() {
	configCmd.PersistentFlags().BoolVar(&affectGlobalConfig, "global", false, "Affect global config")
	configCmd.AddCommand(showCmd)
	configCmd.AddCommand(getCmd)
	configCmd.AddCommand(setCmd)
	rootCmd.AddCommand(configCmd)
}

func doShowConfig() {
	encoder := yaml.NewEncoder(os.Stdout)
	encoder.SetIndent(configuration.INDENTATION)
	encoder.Encode(config)
}

func doGetConfigValue(key string) {
	if affectGlobalConfig {
		assertOperation("retrieving configuration value", configGlobal.ReadField(key, os.Stdout))
	} else {
		assertOperation("retrieving configuration value", config.ReadField(key, os.Stdout))
	}
}

func doSetConfigValue(key, value string) {
	if affectGlobalConfig {
		assertOperation("writing configuration value", configGlobal.WriteField(key, value))
		// TODO: validate configuration settings
		assertOperation("writing configuration file", configGlobal.WriteConfiguration(userConfigRoot, &configSystem))
	} else {
		if projectRoot == "" {
			fmt.Fprintln(os.Stderr, "Called outside of project scope; refusing to modify global configuration unless '--global' is explicitly specified.")
			os.Exit(1)
		}
		assertOperation("writing configuration value", config.WriteField(key, value))
		// TODO: validate configuration settings
		assertOperation("validating application settings", validateAppSettings(true))
		assertOperation("writing configuration file", config.WriteConfiguration(projectRoot, &configGlobal))
	}
}
