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
		Short: "Run make targets",
		Long:  `Execute make targets and build commands`,
	}

	// Add some common flags that make commands might use as persistent flags
	// These will be available to all subcommands

	// Add make target commands
	testCmd := createMakeTargetCommand("test", "Run tests", "Run the test suite")
	lintCmd := createMakeTargetCommand("lint", "Run linter", "Run 'golangci-lint run'")

	// Add subcommands
	rootCmd.AddCommand(ophis.Command(nil), testCmd, lintCmd)
	return rootCmd
}

// createMakeTargetCommand creates a cobra command for a specific make target
func createMakeTargetCommand(target, short, long string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   target,
		Short: short,
		Long:  long,
		RunE: func(cmd *cobra.Command, args []string) error {
			execArgs := []string{target}
			execArgs = append(execArgs, buildArgs(cmd, args)...)
			subCmd := exec.CommandContext(cmd.Context(), "make", execArgs...)
			subCmd.Stdout = cmd.OutOrStdout()
			subCmd.Stderr = cmd.ErrOrStderr()
			return subCmd.Run()
		},
	}

	cmd.Flags().StringP("file", "f", "", "Use FILE as a makefile")
	cmd.Flags().StringP("directory", "C", "", "Change to directory before doing anything")
	err := cmd.MarkFlagRequired("directory")
	if err != nil {
		panic(err)
	}

	return cmd
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
