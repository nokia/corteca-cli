// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package cmd

import (
	"corteca/internal/configuration"
	"corteca/internal/device"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	Use:   "deploy DEVICE",
	Short: "Deploy application",
	Long:  `Deploy application to a specified device`,
	Args:  cobra.ExactArgs(1),
	Run:   func(cmd *cobra.Command, args []string) { doDeployApp(args[0]) },
}

func init() {
	rootCmd.AddCommand(deployCmd)
}

func doDeployApp(deviceName string) {
	requireBuildArtifact()

	deployDevice, found := config.Devices[deviceName]
	if !found {
		failOperation(fmt.Sprintf("device '%v' not found", deviceName))
	}

	// assemble the deploy context
	context := map[string]any{
		"device": configuration.ToDictionary(deployDevice),
		"app":    configuration.ToDictionary(config.App),
		"denv":   denv,
	}
	context["device"].(map[string]any)["name"] = deviceName

	// step 1: connect to the device console
	fmt.Printf("Connecting to device console at %v...\n", deployDevice.Addr)
	conn, err := device.Connect(&deployDevice.Endpoint, config.Deploy.LogFile) // TODO: render logFile before passing
	assertOperation("connecting to device console", err)
	defer conn.Close()

	// step 2: acquire target CPU architecture
	cpuArch, err := device.DiscoverTargetCPUarch(*conn)
	if err != nil {
		assertOperation("discovering device cpu architecture", err)
	}
	fmt.Printf("Discovered %v's CPU arch: \"%v\"\n", deployDevice, cpuArch)

	// step 3: validate existence of build artifact for the acquired architecture
	buildArtifact, ok := buildArtifacts[cpuArch]
	if !ok {
		failOperation(fmt.Sprintf("no build artifact present for target architecture %v", cpuArch))
	}

	// step 4: publish build artifact(s) if a publish target has been specified in the deploy source
	if deployDevice.Source.Publish != "" {
		publishTargetName := deployDevice.Source.Publish
		fmt.Printf("Publishing \"%v\" artifact to \"%v\"\n", cpuArch, publishTargetName)
		doPublishApp(publishTargetName, cpuArch, false)
	}

	// step 5: execute deployment sequence
	fmt.Printf("Deploying %v...\n", buildArtifact)
	context["buildArtifact"] = filepath.Base(buildArtifact)
	// TODO: uninstall application from target if already exists (add cmd to configuration object?)
	assertOperation("executing deployment sequence", conn.ExecuteSequence("deployment", config.Deploy.Sequence, context))
}
