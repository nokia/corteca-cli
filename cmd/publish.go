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
	Use:               "publish TARGET [ARCH]",
	Short:             "Publish application artifact(s) to specified target, optionally filtering by architecture.",
	Long:              "Publish application artifact(s) to specified target, optionally filtering by architecture.",
	Args:              cobra.RangeArgs(1, 2),
	ValidArgsFunction: validPublishArgsFunc,
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
	publishCmd.PersistentFlags().BoolVar(&skipLocalConfig, "global", false, "Affect global config & ignore any project-local configuration")
	publishCmd.PersistentFlags().StringVarP(&specifiedArtifact, "artifact", "a", "", "Specify an artifact in the form of '[ARCH]:imagetype:/path/to/file', architecture=(aarch64|armv7l|x86_64), imagetype=(rootfs|oci)")
	publishCmd.RegisterFlagCompletionFunc("artifact", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"tar.gz"}, cobra.ShellCompDirectiveFilterFileExt
	})
	rootCmd.AddCommand(publishCmd)
}

func doPublishApp(targetName string, arch string, wait bool) {
	requireBuildArtifact()
	if specifiedArtifact != "" {
		if arch != "" && cmdContext.Arch != arch {
			fmt.Printf("Warning: differing architectures [%s,%s] were specified!\nPublishing %s...", arch, cmdContext.Arch, cmdContext.Arch)
		}
		arch = cmdContext.Arch
	}

	target, found := config.Publish[targetName]
	if !found {
		failOperation(fmt.Sprintf("publish target '%s' not found", targetName))
	}

	switch target.Method {
	case configuration.PUBLISH_METHOD_LISTEN:
		handlePublishMethodListen(target, wait)
	case configuration.PUBLISH_METHOD_PUT:
		handlePublishMethodPut(target, arch)
	case configuration.PUBLISH_METHOD_COPY:
		failOperation("not implemented yet")
	case configuration.PUBLISH_METHOD_PUSH:
		handlePublishMethodPush(target, arch)
	case configuration.PUBLISH_METHOD_REGISTRY:
		handlePublishMethodRegistry(target, arch, wait)
	default:
		failOperation(fmt.Sprintf("unknown publish method %v", target.Method))
	}
}

func handlePublishMethodListen(target configuration.PublishTarget, wait bool) {
	doListen(target, wait)
}

func handlePublishMethodRegistry(target configuration.PublishTarget, arch string, wait bool) {
	artifact, found := getArtifact(arch, ociSuffix)
	if !found {
		failOperation(fmt.Sprintf(artifactNotFoundMessage, arch, ociSuffix))
	}
	tag, err := publish.GenerateTag(config.App, artifact, arch)
	assertOperation("generating tag", err)

	registryURL, err := url.Parse(target.Addr)
	assertOperation("parsing registry url", err)

	hostPort := net.JoinHostPort(registryURL.Hostname(), registryURL.Port())
	registryServer, err := publish.StartRegistry(hostPort)
	if err != nil {
		failOperation(fmt.Sprintf("failed to start local registry: %v", err))
	}

	if registryURL.Hostname() == "0.0.0.0" {
		registryURL.Host = net.JoinHostPort("127.0.0.1", registryURL.Port())
	}

	err = publish.PushImage(artifact, registryURL, "", tag, false)
	assertOperation(fmt.Sprintf("pushing image %s to registry", artifact), err)

	if wait {
		waitForInterruptSignal()
		if err := registryServer.Shutdown(context.Background()); err != nil {
			fmt.Printf("failed to shutdown registry server: %v", err)
		}
	} else {
		fmt.Printf("Serving %v on %v\n", hostPort, registryURL.String())
	}
}

func handlePublishMethodPut(target configuration.PublishTarget, arch string) {
	artifact, found := getArtifact(arch, rootfsSuffix)
	if !found {
		failOperation(fmt.Sprintf(artifactNotFoundMessage, arch, rootfsSuffix))
	}

	url, token, err := publish.AuthenticateHttp(target.Addr, target.Auth, target.Token)
	assertOperation("performing http authentication", err)
	doPut(artifact, url, token)
}

func handlePublishMethodPush(target configuration.PublishTarget, arch string) {
	artifact, found := getArtifact(arch, ociSuffix)
	if !found {
		failOperation(fmt.Sprintf(artifactNotFoundMessage, arch, ociSuffix))
	}
	url, token, err := publish.AuthenticateHttp(target.Addr, target.Auth, target.Token)
	assertOperation("performing http authentication", err)

	doPush(artifact, url, token, arch)
}

func doPush(artifact string, url *url.URL, token, arch string) {
	tag, err := publish.GenerateTag(config.App, artifact, arch)
	assertOperation("generating tag", err)
	err = publish.PushImage(artifact, url, token, tag, true)
	assertOperation(fmt.Sprintf("pushing image %s to registry", artifact), err)
}

func getArtifact(arch, suffix string) (string, bool) {
	artifactKey := fmt.Sprintf("%s-%s", arch, suffix)
	artifactFilename, found := cmdContext.BuildArtifacts[artifactKey]
	return artifactFilename, found
}

func doListen(target configuration.PublishTarget, wait bool) {
	u, err := url.Parse(target.Addr)
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

func doPut(artifact string, url *url.URL, token string) {
	if err := publish.HttpPut(artifact, *url, token); err != nil {
		assertOperation(fmt.Sprintf("while uploading file \"%s\" with HTTP(S) PUT", artifact), err)
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
		architectures := make([]string, 0, len(config.Build.Toolchains.Architectures))
		for k := range config.Build.Toolchains.Architectures {
			if strings.HasPrefix(k, toComplete) {
				architectures = append(architectures, k)
			}
		}
		return architectures, cobra.ShellCompDirectiveNoFileComp
	}

	return nil, cobra.ShellCompDirectiveNoFileComp
}
