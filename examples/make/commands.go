package main

import (
	"os/exec"

	"github.com/njayp/ophis/bridge"
	"github.com/njayp/ophis/commands"
	"github.com/spf13/cobra"
)

func createMakeCommands() *cobra.Command {
	// Create the root make command
	rootCmd := &cobra.Command{
		Use:   "make",
		Short: "Run make commands",
		Long:  `Execute make targets and build commands`,
	}

	mcpCmd := commands.MCPCommand(&CommandFactory{
		rootCmd: rootCmd,
	}, &bridge.MCPCommandConfig{
		AppName:    AppName,
		AppVersion: AppVersion,
	})

	// Add some common flags that make commands might use as persistent flags
	// These will be available to all subcommands
	rootCmd.PersistentFlags().StringP("file", "f", "", "Use FILE as a makefile")
	rootCmd.PersistentFlags().StringP("directory", "C", "", "Change to directory before doing anything")
	rootCmd.PersistentFlags().BoolP("dry-run", "n", false, "Don't actually run any commands; just print them")
	rootCmd.PersistentFlags().BoolP("silent", "s", false, "Don't print the commands as they are executed")

	// Add make target commands
	testCmd := createMakeTargetCommand("test", "Run tests", "Run the test suite using 'make test'")
	lintCmd := createMakeTargetCommand("lint", "Run linter", "Run 'golangci-lint run' using 'make test'")

	// Add subcommands
	rootCmd.AddCommand(mcpCmd, testCmd, lintCmd)
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	return rootCmd
}

// createMakeTargetCommand creates a cobra command for a specific make target
func createMakeTargetCommand(target, short, long string) *cobra.Command {
	return &cobra.Command{
		Use:   target,
		Short: short,
		Long:  long,
		Run: func(cmd *cobra.Command, args []string) {
			makeArgs := []string{}

			// Add flags to make arguments
			makeArgs = appendMakeFlags(cmd, makeArgs)

			// Add target
			makeArgs = append(makeArgs, target)

			// Add any additional positional arguments
			makeArgs = append(makeArgs, args...)

			data, _ := exec.CommandContext(cmd.Context(), "make", makeArgs...).CombinedOutput()
			cmd.Print(string(data))
		},
	}
}

// appendMakeFlags converts cobra flags to make command arguments
func appendMakeFlags(cmd *cobra.Command, args []string) []string {
	flags := cmd.Flags()

	// Handle --file/-f flag
	if file, _ := flags.GetString("file"); file != "" {
		args = append(args, "-f", file)
	}

	// Handle --directory/-C flag
	if directory, _ := flags.GetString("directory"); directory != "" {
		args = append(args, "-C", directory)
	}

	// Handle --dry-run/-n flag
	if dryRun, _ := flags.GetBool("dry-run"); dryRun {
		args = append(args, "-n")
	}

	// Handle --silent/-s flag
	if silent, _ := flags.GetBool("silent"); silent {
		args = append(args, "-s")
	}

	return args
}
