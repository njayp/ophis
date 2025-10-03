package main

import (
	"os"
	"os/exec"

	"github.com/njayp/ophis"
	"github.com/spf13/cobra"
)

func main() {
	if err := makeCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

type flags struct {
	File      string
	Directory string
}

func makeCmd() *cobra.Command {
	flags := &flags{}
	// Create the root make command
	cmd := &cobra.Command{
		Use:   "make [targets...]",
		Short: "Run make targets",
		Long:  `Execute make targets`,
		RunE:  flags.run,
		Args:  cobra.ArbitraryArgs,
	}

	cmd.Flags().StringVarP(&flags.File, "file", "f", "", "Use FILE as a makefile")
	cmd.Flags().StringVarP(&flags.Directory, "directory", "C", "", "Change to directory before doing anything")
	err := cmd.MarkFlagRequired("directory")
	if err != nil {
		panic(err)
	}

	// Add subcommands
	cmd.AddCommand(ophis.Command(nil))
	return cmd
}

func (f *flags) run(cmd *cobra.Command, args []string) error {
	// Handle --file/-f flag
	if f.File != "" {
		args = append(args, "-f", f.File)
	}

	// Handle --directory/-C flag
	if f.Directory != "" {
		args = append(args, "-C", f.Directory)
	}

	subCmd := exec.CommandContext(cmd.Context(), "make", args...)
	subCmd.Stdout = cmd.OutOrStdout()
	subCmd.Stderr = cmd.ErrOrStderr()
	return subCmd.Run()
}
