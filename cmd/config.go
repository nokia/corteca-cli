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

var getCmd = &cobra.Command{
	Use:   "get key",
	Short: "Read a configuration value",
	Long:  "Read a configuration value",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := ""
		if len(args) > 0 {
			key = args[0]
		}
		doGetConfigValue(key)
	},
}

var setCmd = &cobra.Command{
	Use:   "set key value",
	Short: "Set a configuration value",
	Long:  "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	Run:   func(cmd *cobra.Command, args []string) { doSetConfigValue(args[0], args[1], false) },
}

var addCmd = &cobra.Command{
	Use:   "add key value",
	Short: "Add (append) a configuration value",
	Long:  "Add (append) a configuration value",
	Args:  cobra.ExactArgs(2),
	Run:   func(cmd *cobra.Command, args []string) { doSetConfigValue(args[0], args[1], true) },
}

func init() {
	configCmd.PersistentFlags().BoolVar(&affectGlobalConfig, "global", false, "Affect global config")
	configCmd.AddCommand(getCmd)
	configCmd.AddCommand(setCmd)
	configCmd.AddCommand(addCmd)
	rootCmd.AddCommand(configCmd)
}

func doGetConfigValue(key string) {
	var (
		field any
		err   error
	)

	if affectGlobalConfig {
		field, err = configGlobal.ReadField(key)
	} else {
		field, err = config.ReadField(key)
	}
	assertOperation("retrieving config value", err)
	enc := yaml.NewEncoder(os.Stdout)
	enc.SetIndent(configuration.INDENTATION)
	enc.Encode(field)
}

func doSetConfigValue(key, value string, append bool) {
	if affectGlobalConfig {
		assertOperation("writing configuration value", configGlobal.WriteField(key, value, append))
		// TODO: validate configuration settings
		assertOperation("writing configuration file", configGlobal.WriteConfiguration(userConfigRoot, &configSystem))
	} else {
		if projectRoot == "" {
			fmt.Fprintln(os.Stderr, "Called outside of project scope; refusing to modify global configuration unless '--global' is explicitly specified.")
			os.Exit(1)
		}
		assertOperation("writing configuration value", config.WriteField(key, value, append))
		// TODO: validate configuration settings
		assertOperation("validating application settings", validateAppSettings(false))
		assertOperation("writing configuration file", config.WriteConfiguration(projectRoot, &configGlobal))
	}
}
