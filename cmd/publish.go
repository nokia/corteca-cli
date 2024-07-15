// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package cmd

import (
	"context"
	"corteca/internal/configuration"
	"corteca/internal/publish"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

var publishCmd = &cobra.Command{
	Use:   "publish TARGET [ARCH]",
	Short: "Publish application artifact(s) to specified target, optionally filtering by architecture.",
	Long:  "Publish application artifact(s) to specified target, optionally filtering by architecture.",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		targetName := args[0]
		arch := ""

		if len(args) > 1 {
			arch = args[1]
		}

		doPublishApp(targetName, arch, true)
	},
}

func init() {
	rootCmd.AddCommand(publishCmd)
}

func doPublishApp(targetName string, arch string, wait bool) {
	requireBuildArtifact()

	target, found := config.Publish[targetName]
	if !found {
		failOperation(fmt.Sprintf("publish target '%s' not found", targetName))
	}

	if _, found = cmdContext.BuildArtifacts[arch]; !found && arch != "" {
		failOperation(fmt.Sprintf("No build artifact found for specified architecture \"%v\"", arch))
	}

	switch target.Method {
	case configuration.PUBLISH_METHOD_LISTEN:
		doListen(target, wait)

	case configuration.PUBLISH_METHOD_PUT:

		url, token, err := publish.AuthenticateHttp(target)
		assertOperation("performing http authentication", err)

		doPut(arch, url, token)

	case configuration.PUBLISH_METHOD_COPY:
		failOperation("not implemented yet")
	default:
		failOperation(fmt.Sprintf("unknown publish method %v", target.Method))
	}
}

func doListen(target configuration.PublishTarget, wait bool) {

	u, err := url.Parse(target.Addr)
	assertOperation("parsing target url", err)

	serverRoot := distFolder
	srv, err := publish.ListenAsync(serverRoot, u)
	assertOperation("starting server", err)
	if wait {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		fmt.Printf("Serving %v on %v; press Ctrl-C to stop...\n", serverRoot, u.String())
		<-sigChan // wait for signal
		srv.Shutdown(context.Background())
	} else {
		fmt.Printf("Serving %v on %v\n", serverRoot, u.String())
	}
}

func doPut(arch string, url *url.URL, token string) {

	for artifactArch, artifactBinaryFile := range cmdContext.BuildArtifacts {
		if arch == "" || artifactArch == arch {
			if err := publish.HttpPut(artifactBinaryFile, *url, token); err != nil {
				assertOperation(fmt.Sprintf("while uploading file \"%s\" with HTTP(S) PUT", artifactBinaryFile), err)
			}
		}
	}
}
