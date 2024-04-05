package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var execCmd = &cobra.Command{
	Use:   "exec NAMED-SEQUENCE DEVICE",
	Short: "Execute sequence",
	Long:  `Execute sequence to a specified device`,
	Args:  cobra.ExactArgs(2),
	Run:   func(cmd *cobra.Command, args []string) { doExecSequence(args[0], args[1]) },
}

func init() {
	rootCmd.AddCommand(execCmd)
}

func doExecSequence(sequence, deviceName string) {
	if sequence == "deploy" {
		doDeployApp(deviceName)
	} else {
		failOperation(fmt.Sprintf("Sequence '%v' not supported yet", sequence))
	}
}
