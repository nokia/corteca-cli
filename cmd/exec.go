package cmd

import (
	"corteca/internal/configuration"
	"corteca/internal/device"
	"corteca/internal/platform"
	"corteca/internal/tui"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var execCmd = &cobra.Command{
	Use:               "exec NAMED-SEQUENCE DEVICE",
	Short:             "Execute sequence",
	Long:              `Execute sequence to a specified device`,
	Example:           "",
	Args:              cobra.ExactArgs(2),
	ValidArgsFunction: validExecArgsFunc,
	Run:               func(cmd *cobra.Command, args []string) { doExecSequence(args[0], args[1]) },
}

var logFile string
var publishTargetName string

func init() {
	execCmd.PersistentFlags().StringVarP(&specifiedArtifact, "artifact", "a", "", "Specify an artifact in the form of 'architecture:imagetype:/path/to/file', architecture=(aarch64|armv7l|x86_64), imagetype=(rootfs|oci)")
	execCmd.RegisterFlagCompletionFunc("artifact", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"tar.gz, tar"}, cobra.ShellCompDirectiveFilterFileExt
	})
	rootCmd.AddCommand(execCmd)
	execCmd.PersistentFlags().StringVar(&logFile, "logfile", platform.DefaultLog, "Specify where SSH logs will be stored")
	execCmd.PersistentFlags().StringVar(&publishTargetName, "publish", "", "Publish application artifact to specified target")
	execCmd.PersistentFlags().BoolVar(&skipLocalConfig, "global", false, "Affect global config & ignore any project-local configuration")
}

func doExecSequence(sequence, deviceName string) {
	if _, exists := config.Sequences[sequence]; !exists {
		failOperation(fmt.Sprintf("Sequence '%s' not supported yet", sequence))
	}

	requireBuildArtifact()
	var found bool
	configuration.CmdContext.Device.Name = deviceName
	configuration.CmdContext.Device.DeployDevice, found = config.Devices[deviceName]
	if !found {
		failOperation(fmt.Sprintf("device '%s' not found", deviceName))
	}

	// connect to the device console
	dev, err := device.NewDevice(configuration.CmdContext.Device.Endpoint, logFile)
	if err != nil {
		failOperation(fmt.Sprintf("could not create device %s", deviceName))
	}
	dispatcher, err := dev.Connect()
	assertOperation("connecting to device", err)
	defer dev.Close()

	if publishTargetName != "" {
		if dev.GetProtocol() == device.ConnectionSSH {
			containerType := device.DetectContainerFramework(dispatcher)
			if containerType == "" {
				failOperation("no valid container framework found on device")
			}
			configuration.CmdContext.Build.Options.OutputType = containerType
		} else {
			configuration.CmdContext.Build.Options.OutputType = "oci"
		}
	}

	// populate contextCmd
	configuration.CmdContext.Arch, err = dev.DiscoverTargetCPUArch(dispatcher)
	if err != nil {
		assertOperation("discovering device cpu architecture", err)
	}
	

	artifactKey := fmt.Sprintf("%s-%s", configuration.CmdContext.Arch, configuration.CmdContext.Build.Options.OutputType)
	buildArtifact, ok := configuration.CmdContext.BuildArtifacts[artifactKey]
	if !ok {
		failOperation(fmt.Sprintf("no build artifact present for target architecture \"%s\"", configuration.CmdContext.Arch))
	}

	configuration.CmdContext.BuildArtifact = filepath.Base(buildArtifact)
	configuration.CmdContext.Publish.PublishTarget = config.Publish[publishTargetName]
	configuration.CmdContext.Publish.Name = publishTargetName
	// publish build artifact(s) if a publish target has been specified in the deploy source
	if publishTargetName != "" {
		tui.LogNormal("Publishing \"%s\" artifact to \"%s\"", configuration.CmdContext.Arch, configuration.CmdContext.Publish.Name)
		doPublishApp(configuration.CmdContext.Publish.Name, configuration.CmdContext.Arch, false)
	}

	// execute the sequence
	tui.LogNormal("Deploying %s...", buildArtifact)
	if err = config.Sequences.Execute(dispatcher, sequence); err != nil {
		tui.LogError("Error while %v: %v", "executing "+sequence+" sequence", err.Error())
		return
	}
	tui.DisplaySuccessMsg("Sequence completed successfully!")
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
