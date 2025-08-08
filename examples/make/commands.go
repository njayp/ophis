// Package main provides an example MCP server that exposes make commands.
// This demonstrates how to use njayp/ophis to convert a make-based build system into an MCP server.
package main

import (
	"os/exec"

	"github.com/njayp/ophis"
	"github.com/spf13/cobra"
)

func createMakeCommands() *cobra.Command {
	// Create the root make command
	rootCmd := &cobra.Command{
		Use:   "make",
		Short: "Run make commands",
		Long:  `Execute make targets and build commands`,
	}

	mcpCmd := ophis.Command(nil)

	// Add some common flags that make commands might use as persistent flags
	// These will be available to all subcommands
	rootCmd.PersistentFlags().StringP("file", "f", "", "Use FILE as a makefile")
	rootCmd.PersistentFlags().StringP("directory", "C", "", "Change to directory before doing anything")

	// Add make target commands
	testCmd := createMakeTargetCommand("test", "Run tests", "Run the test suite using 'make test'")
	lintCmd := createMakeTargetCommand("lint", "Run linter", "Run 'golangci-lint run' using 'make test'")

	// Add subcommands
	rootCmd.AddCommand(mcpCmd, testCmd, lintCmd)
	return rootCmd
}

// createMakeTargetCommand creates a cobra command for a specific make target
func createMakeTargetCommand(target, short, long string) *cobra.Command {
	return &cobra.Command{
		Use:   target,
		Short: short,
		Long:  long,
		RunE: func(cmd *cobra.Command, args []string) error {
			execArgs := []string{target}
			execArgs = append(execArgs, buildArgs(cmd, args)...)
			data, err := exec.CommandContext(cmd.Context(), "make", execArgs...).CombinedOutput()
			cmd.Print(string(data))
			return err
		},
	}
}

func buildArgs(cmd *cobra.Command, args []string) []string {
	flags := cmd.Flags()

	// Handle --file/-f flag
	if file, _ := flags.GetString("file"); file != "" {
		args = append(args, "-f", file)
	}

	// Handle --directory/-C flag
	if directory, _ := flags.GetString("directory"); directory != "" {
		args = append(args, "-C", directory)
	}

	return args
}
