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
	var found bool
	cmdContext.Device.DeployDevice, found = config.Devices[deviceName]
	cmdContext.Device.Name = deviceName
	if !found {
		failOperation(fmt.Sprintf("device '%v' not found", deviceName))
	}

	// step 1: connect to the device console
	fmt.Printf("Connecting to device console at %v...\n", cmdContext.Device.Addr)
	conn, err := device.Connect(&cmdContext.Device.Endpoint, config.Deploy.LogFile) // TODO: render logFile before passing
	assertOperation("connecting to device console", err)
	defer conn.Close()

	// step 2: acquire target CPU architecture
	cmdContext.Arch, err = device.DiscoverTargetCPUarch(*conn)
	if err != nil {
		assertOperation("discovering device cpu architecture", err)
	}
	fmt.Printf("Discovered CPU architecture for '%s': '%s'\n", deviceName, cmdContext.Arch)

	// step 3: validate existence of build artifact for the acquired architecture
	buildArtifact, ok := cmdContext.BuildArtifacts[cmdContext.Arch]
	if !ok {
		failOperation(fmt.Sprintf("no build artifact present for target architecture %v", cmdContext.Arch))
	}

	// step 4: publish build artifact(s) if a publish target has been specified in the deploy source
	if cmdContext.Device.Source.Publish != "" {
		cmdContext.Source.DownloadSource = cmdContext.Device.Source
		cmdContext.Source.Name = cmdContext.Device.Source.Publish
		fmt.Printf("Publishing \"%v\" artifact to \"%v\"\n", cmdContext.Arch, cmdContext.Source.Name)
		doPublishApp(cmdContext.Source.Name, cmdContext.Arch, false)
	}

	// step 5: execute deployment sequence
	fmt.Printf("Deploying %v...\n", buildArtifact)
	cmdContext.BuildArtifact = filepath.Base(buildArtifact)
	// TODO: uninstall application from target if already exists (add cmd to configuration object?)
	assertOperation("executing deployment sequence", conn.ExecuteSequence("deployment", config.Deploy.Sequence, configuration.ToDictionary(cmdContext)))
}
