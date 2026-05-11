package cmd

import (
	"github.com/nokia/corteca-cli/internal/configuration"
	"github.com/nokia/corteca-cli/internal/device"
	_ "github.com/nokia/corteca-cli/internal/device/cwmp"
	_ "github.com/nokia/corteca-cli/internal/device/ssh"
	"github.com/nokia/corteca-cli/internal/platform"
	"github.com/nokia/corteca-cli/internal/tui"
	"fmt"
	"io"
	"os"
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
	execCmd.RegisterFlagCompletionFunc("artifact", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"tar.gz, tar"}, cobra.ShellCompDirectiveFilterFileExt
	})
	rootCmd.AddCommand(execCmd)
	execCmd.PersistentFlags().StringVar(&logFile, "logfile", platform.DefaultLog, "Specify where SSH logs will be stored")
	execCmd.PersistentFlags().StringVar(&publishTargetName, "publish", "", "Publish application artifact to specified target")
	execCmd.PersistentFlags().StringVarP(&artifact, "artifact", "a", "", "Specify the path to a an artifact to publish")
	execCmd.PersistentFlags().BoolVar(&skipLocalConfig, "global", false, "Affect global config & ignore any project-local configuration")
}

func doExecSequence(sequencename, deviceName string) {
	if devConfig, found := config.Devices[deviceName]; !found {
		failOperation(fmt.Sprintf("no config for device '%s' was found", deviceName))
	} else {
		if !skipLocalConfig {
			requireBuildArtifact()
		}
		configuration.GetCmdContext().Device.DeviceConfig = devConfig
		configuration.GetCmdContext().Device.Name = deviceName
		configuration.GetCmdContext().Arch = configuration.GetCmdContext().Device.Architecture
	}

	// prepare log file
	var log io.WriteCloser
	switch strings.ToLower(logFile) {
	case "stdout":
		log = os.Stdout
	case "stderr":
		log = os.Stderr
	default:
		if f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666); err != nil {
			failOperation(fmt.Sprintf("Could not create log file: %s", err.Error()))
		} else {
			log = f
			defer func() {
				if err := f.Close(); err != nil {
					tui.LogError("could not close log file (%s)", err.Error())
				}
			}()
		}
	}

	// connect to the device console
	device, err := device.NewDevice(&configuration.GetCmdContext().Device.DeviceConfig, log)
	if err != nil {
		failOperation(fmt.Sprintf("could not create device %s (%s)", deviceName, err.Error()))
	}
	tui.LogNormal("Selected device '%s', protocol: %s", deviceName, device.GetProtocol())
	defer device.Close()

	// publish build artifact(s) if a publish target has been specified in the deploy source
	if publishTargetName != "" {
		configuration.GetCmdContext().Publish.PublishTarget = config.Publish[publishTargetName]
		configuration.GetCmdContext().Publish.Name = publishTargetName
		tui.LogNormal("Publishing artifact to '%s'", configuration.GetCmdContext().Publish.Name)
		doPublishApp(configuration.GetCmdContext().Publish.Name, false)
	}

	// execute the sequence
	if err = config.Sequences.Execute(device, sequencename); err != nil {
		tui.LogError("Error while executing sequence '%s': %s", sequencename, err.Error())
	} else {
		tui.DisplaySuccessMsg("Sequence completed successfully!")
	}
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
