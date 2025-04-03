package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var writeToFileErr error

var docCmd = &cobra.Command{
	Use:     "doc COMMAND(s)",
	Short:   "Generate documentation for corteca commands",
	Long:    "Generate documentation for corteca commands\nIf no specific command is given, documentation will be generated for every command",
	Example: "",
	Args: func(cmd *cobra.Command, args []string) error {
		// Get the number of subcommands in the root command dynamically
		return cobra.MaximumNArgs(len(rootCmd.Commands()))(cmd, args)
	},
	ValidArgsFunction: validDocArgsFunc,
	Run:               func(cmd *cobra.Command, args []string) { doGenerateDocumentation(args) },
}

func init() {
	rootCmd.AddCommand(docCmd)
}

// Auto-generate cortecacli's commands documentation
func doGenerateDocumentation(cmdNames []string) {
	workingDir, err := os.Getwd()
	if err != nil {
		failOperation(err.Error())
	}
	if strings.ToLower(filepath.Base(workingDir)) != "cortecacli" {
		failOperation("execution of 'corteca doc' command must be inside cortecacli's directory")
	}
	// Generate documentation for cortecacli's commands
	if len(cmdNames) == 0 {
		for _, cmd := range rootCmd.Commands() {
			if !cmd.IsAvailableCommand() || cmd.IsAdditionalHelpTopicCommand() {
				continue
			}
			cmdFile := createMdFile(cmd)
			defer cmdFile.Close()
			err = generateMarkdown(cmd, cmdFile, false)
			assertOperation("generating documentation for corteca commands", err)
		}
		return
	}
	// Generate documentation for specific commands
	initialCmdNamesSize := len(cmdNames)
	notMatchedCmds := []string{}
	for index := 0; index < len(cmdNames); index++ {
		givenCmdName := cmdNames[index]
		for _, cmd := range rootCmd.Commands() {
			// If any of the commands' name matches the given command's name, then the command exist and we generate the md file
			if cmd.Name() == givenCmdName {
				// Removing the matched command from the search list and decrease index to match the new array size
				cmdNames = append(cmdNames[:index], cmdNames[index+1:]...)
				index--

				cmdFile := createMdFile(cmd)
				defer cmdFile.Close()
				err = generateMarkdown(cmd, cmdFile, false)
				assertOperation("generating documentation for specific corteca command(s)", err)
			}
		}
		// If size of cmdNames slice has not been reduced, then the command wasn't matched with a cortecacli command
		if len(cmdNames) == initialCmdNamesSize {
			notMatchedCmds = append(notMatchedCmds, givenCmdName)
		}
	}
	// If size is different that means at least one documentation file has been generated
	if len(cmdNames) != initialCmdNamesSize {
		if len(notMatchedCmds) == 0 {
			fmt.Println("Cortecacli's commands documentation has generated successfuly!")
		} else {
			fmt.Fprintf(os.Stdout, "The following commands are not part of cortecacli: %s\n", notMatchedCmds)
		}
	} else {
		fmt.Fprintf(os.Stdout, "No command documentation was generated\nThe following commands are not part of cortecacli: %s\n", notMatchedCmds)
	}
}

func generateMarkdown(cmd *cobra.Command, cmdFile *os.File, isSubcmd bool) error {
	// Write command usage
	if isSubcmd {
		_, writeToFileErr = cmdFile.WriteString(fmt.Sprintf("### Corteca %s %s\n\n%s\n\n", cmd.Parent().Name(), cmd.Name(), cmd.Long))
	} else {
		_, writeToFileErr = cmdFile.WriteString(fmt.Sprintf("## Corteca %s\n\n%s\n\n", cmd.Name(), cmd.Long))
	}
	if writeToFileErr != nil {
		return writeToFileErr
	}
	cmdFile.WriteString(fmt.Sprintf("### Usage:\n\n```\n%s\n```\n\n", cmd.UseLine()))
	cmdFile.WriteString(fmt.Sprintf("### Flags:\n\n```\n%s%s\n```\n\n", cmd.InheritedFlags().FlagUsages(), cmd.LocalFlags().FlagUsages()))
	cmdFile.WriteString(fmt.Sprintf("### Example:\n\n```\n%s\n```\n\n", cmd.Example))

	// Append subcommands to parent command's documentation, if any
	if len(cmd.Commands()) > 0 {
		cmdFile.WriteString("## Subcommands:\n\n")
		for _, subCmd := range cmd.Commands() {
			// Recursive call for nested subcommands
			err := generateMarkdown(subCmd, cmdFile, true)
			assertOperation("generating documentation for corteca subcommands", err)
		}
	}

	cmdFile.WriteString("\n")
	return nil
}

func createMdFile(cmd *cobra.Command) *os.File {
	cmdFile, err := os.Create(fmt.Sprintf("./doc/reference/corteca_%s.md", cmd.Name()))
	if err != nil {
		failOperation(err.Error())
	}
	return cmdFile
}

func validDocArgsFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	cmds := make([]string, 0, len(rootCmd.Commands()))
	for _, cmd := range rootCmd.Commands() {
		if strings.HasPrefix(cmd.Name(), toComplete) && cmd.Name() != "__complete" {
			cmds = append(cmds, cmd.Name())
		}
	}
	return cmds, cobra.ShellCompDirectiveNoFileComp
}
