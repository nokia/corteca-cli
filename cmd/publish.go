// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package cmd

import (
	"context"
	"corteca/internal/configuration"
	"corteca/internal/publish"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
)

const (
	ociSuffix               = "oci"
	rootfsSuffix            = "rootfs"
	artifactNotFoundMessage = "No build artifact found for [%s,%s]"
)

var publishCmd = &cobra.Command{
	Use:               "publish TARGET",
	Short:             "Publish application artifact(s) to specified target, optionally filtering by architecture.",
	Long:              "Publish application artifact(s) to specified target, optionally filtering by architecture.",
	Example:           "",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: validPublishArgsFunc,
	Run: func(cmd *cobra.Command, args []string) {
		targetName := args[0]
		doPublishApp(targetName, true)
	},
}

func init() {
	publishCmd.PersistentFlags().BoolVar(&skipLocalConfig, "global", false, "Affect global config & ignore any project-local configuration")
	publishCmd.PersistentFlags().StringVarP(&artifact, "artifact", "a", "", "Specify the path to a an artifact to publish")
	publishCmd.RegisterFlagCompletionFunc("artifact", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"tar.gz"}, cobra.ShellCompDirectiveFilterFileExt
	})
	rootCmd.AddCommand(publishCmd)
}

func doPublishApp(targetName string, wait bool) {
	requireBuildArtifact()
	target, found := config.Publish[targetName]
	if !found {
		failOperation(fmt.Sprintf("publish target '%s' not found", targetName))
	}

	switch target.Method {
	case "listen":
		serverConfig := configuration.HttpServerEndpoint{}
		target.Decode(&serverConfig)
		handleListen(serverConfig, wait)
	case "put":
		clientConfig := configuration.HttpClientEndpoint{}
		target.Decode(&clientConfig)
		handlePut(clientConfig, artifact)
	case "push":
		clientConfig := configuration.HttpClientEndpoint{}
		target.Decode(&clientConfig)
		handlePush(clientConfig, artifact)
	case "registry-v2":
		serverConfig := configuration.HttpServerEndpoint{}
		target.Decode(&serverConfig)
		handleRegistry(serverConfig, wait)
	default:
		failOperation(fmt.Sprintf("unknown publish method '%v'", target.Method))
	}
}

func handleListen(target configuration.HttpServerEndpoint, wait bool) {
	u, err := url.Parse(target.Addr.String())
	assertOperation("parsing target url", err)

	serverRoot := distFolder
	srv, err := publish.ListenAsync(serverRoot, u)
	assertOperation("starting server", err)
	if wait {
		waitForInterruptSignal()
		srv.Shutdown(context.Background())
	} else {
		fmt.Printf("Serving %v on %v\n", serverRoot, u.String())
	}
}

func handlePut(target configuration.HttpClientEndpoint, artifact string) {
	// TODO: replace this with target.NewHttpClient() method
	url, err := publish.AuthenticateHttp(target)
	assertOperation("performing http authentication", err)
	if err := publish.HttpPut(artifact, *url, target.Token.String()); err != nil {
		assertOperation(fmt.Sprintf("while uploading file \"%s\" with HTTP(S) PUT", artifact), err)
	}
}

func handlePush(target configuration.HttpClientEndpoint, artifact string) {
	url, err := publish.AuthenticateHttp(target)
	assertOperation("performing http authentication", err)
	err = publish.PushImage(artifact, url, target.Token.String(), true)
	assertOperation(fmt.Sprintf("pushing image %s to registry", artifact), err)
}

func connectableHostPort(host, port string) string {
	switch host {
	case "0.0.0.0":
	case "localhost":
		return net.JoinHostPort("127.0.0.1", port)
	}
	return net.JoinHostPort("127.0.0.1", port)
}

func handleRegistry(target configuration.HttpServerEndpoint, wait bool) {
	registryServer, err := publish.StartRegistry(target.Addr.String(), artifact)
	if err != nil {
		failOperation(fmt.Sprintf("failed to start local registry: %v", err))
	}

	if host, port, err := net.SplitHostPort(registryServer.Addr); err != nil {
		failOperation(fmt.Sprintf("cannot determine registry server addr to connect: %s", err.Error()))
	} else {
		// TODO: below is ugly but will be fixed once publish.PushImage() accepts configuration.HttpClientEndpoint
		url, _ := url.Parse(fmt.Sprintf("https://%s", connectableHostPort(host, port)))
		err = publish.PushImage(artifact, url, "", false)
		assertOperation(fmt.Sprintf("pushing image %s to registry", artifact), err)
	}

	if wait {
		waitForInterruptSignal()
		if err := registryServer.Shutdown(context.Background()); err != nil {
			fmt.Printf("failed to shutdown registry server: %v", err)
		}
	} else {
		fmt.Printf("Serving on %v...\n", registryServer.Addr)
	}
}

func waitForInterruptSignal() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	fmt.Println("Press Ctrl+C to stop...")
	<-sigChan
}

func validPublishArgsFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) == 0 {
		publishTargets := make([]string, 0, len(config.Publish))
		for k := range config.Publish {
			if strings.HasPrefix(k, toComplete) {
				publishTargets = append(publishTargets, k)
			}
		}
		return publishTargets, cobra.ShellCompDirectiveNoFileComp
	} else if len(args) == 1 {
		architectures := make([]string, 0, len(config.Build.Architectures))
		for k := range config.Build.Architectures {
			if strings.HasPrefix(k, toComplete) {
				architectures = append(architectures, k)
			}
		}
		return architectures, cobra.ShellCompDirectiveNoFileComp
	}

	return nil, cobra.ShellCompDirectiveNoFileComp
}
