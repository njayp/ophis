package vscode

import (
	"github.com/spf13/cobra"
)

// Command creates a new Cobra command that manages VSCode MCP configuration
// This command can be added as a subcommand to any Cobra-based application
func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vscode",
		Short: "Configure VSCode MCP servers",
		Long:  "Configure MCP servers for Visual Studio Code",
	}

	// Add subcommands
	cmd.AddCommand(enableCommand(), disableCommand(), listCommand())
	return cmd
}
