package cmd

import (
	"corteca/internal/configuration"
	"corteca/internal/device"
	"fmt"
	"strings"

	"path/filepath"

	"github.com/spf13/cobra"
)

var execCmd = &cobra.Command{
	Use:               "exec NAMED-SEQUENCE DEVICE",
	Short:             "Execute sequence",
	Long:              `Execute sequence to a specified device`,
	Args:              cobra.ExactArgs(2),
	ValidArgsFunction: validExecArgsFunc,
	Run:               func(cmd *cobra.Command, args []string) { doExecSequence(args[0], args[1]) },
}

var sshLogging string
var publishTargetName string

func init() {
	execCmd.PersistentFlags().StringVarP(&specifiedArtifact, "artifact", "a", "", "Specify an artifact in the form of 'architecture:imagetype:/path/to/file', architecture=(aarch64|armv7l|x86_64), imagetype=(rootfs|oci)")
	execCmd.RegisterFlagCompletionFunc("artifact", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"tar.gz, tar"}, cobra.ShellCompDirectiveFilterFileExt
	})
	rootCmd.AddCommand(execCmd)
	execCmd.PersistentFlags().StringVar(&sshLogging, "ssh-log", "/dev/null", "Specify where SSH logs will be stored")
	execCmd.PersistentFlags().StringVar(&publishTargetName, "publish", "", "Publish application artifact to specified target")
	execCmd.PersistentFlags().BoolVar(&skipLocalConfig, "global", false, "Affect global config & ignore any project-local configuration")
}

func doExecSequence(sequence, deviceName string) {
	if _, exists := config.Sequences[sequence]; !exists {
		failOperation(fmt.Sprintf("Sequence '%s' not supported yet", sequence))
	}

	requireBuildArtifact()
	var found bool
	cmdContext.Device.Name = deviceName
	cmdContext.Device.DeployDevice, found = config.Devices[deviceName]
	if !found {
		failOperation(fmt.Sprintf("device '%s' not found", deviceName))
	}

	// connect to the device console
	fmt.Printf("Connecting to device console at %s...\n", cmdContext.Device.Addr)
	conn, err := device.Connect(cmdContext.Device.Endpoint.Addr, cmdContext.Device.Endpoint.Auth, cmdContext.Device.Endpoint.PrivateKeyFile, cmdContext.Device.Endpoint.Password2, sshLogging)
	assertOperation("connecting to device console", err)
	defer conn.Close()

	if publishTargetName != "" {
		containerType := device.ContainerFrameworkType(*conn)
		if containerType == "" {
			failOperation("no valid container framework found on device")
		}
		cmdContext.Build.Options.OutputType = containerType
	}

	// populate contextCmd
	cmdContext.Arch, err = device.DiscoverTargetCPUarch(*conn)
	if err != nil {
		assertOperation("discovering device cpu architecture", err)
	}
	fmt.Printf("Discovered CPU architecture for '%s': '%s'\n", deviceName, cmdContext.Arch)

	artifactKey := fmt.Sprintf("%s-%s", cmdContext.Arch, cmdContext.Build.Options.OutputType)
	buildArtifact, ok := cmdContext.BuildArtifacts[artifactKey]
	if !ok {
		failOperation(fmt.Sprintf("no build artifact present for target architecture \"%s\"", cmdContext.Arch))
	}
	cmdContext.BuildArtifact = filepath.Base(buildArtifact)

	cmdContext.Publish.PublishTarget = config.Publish[publishTargetName]
	cmdContext.Publish.Name = publishTargetName
	// publish build artifact(s) if a publish target has been specified in the deploy source
	if publishTargetName != "" {
		fmt.Printf("Publishing \"%s\" artifact to \"%s\"\n", cmdContext.Arch, cmdContext.Publish.Name)
		doPublishApp(cmdContext.Publish.Name, cmdContext.Arch, false)
	}

	// execute the sequence
	fmt.Printf("Deploying %s...\n", buildArtifact)
	assertOperation("executing "+sequence+" sequence", config.ExecuteSequence(sequence, configuration.ToDictionary(cmdContext), func(cmd string) error {
		_, _, err := conn.SendCmd(cmd)
		return err
	}))
}

func validExecArgsFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) == 0 {
		sequences := make([]string, 0, len(config.Sequences))
		for k := range config.Sequences {
			if strings.HasPrefix(k, toComplete) {
				sequences = append(sequences, k)
			}
		}
		return sequences, cobra.ShellCompDirectiveNoFileComp
	} else if len(args) == 1 {
		devices := make([]string, 0, len(config.Devices))
		for k := range config.Devices {
			if strings.HasPrefix(k, toComplete) {
				devices = append(devices, k)
			}
		}
		return devices, cobra.ShellCompDirectiveNoFileComp
	}

	return nil, cobra.ShellCompDirectiveNoFileComp
}
