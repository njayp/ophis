package vscode

import (
	"github.com/spf13/cobra"
)

// Command creates a new Cobra command for managing VSCode MCP servers.
func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vscode",
		Short: "Manage VSCode MCP servers",
		Long:  "Manage MCP server configuration for Visual Studio Code",
	}

	// Add subcommands
	cmd.AddCommand(enableCommand(), disableCommand(), listCommand())
	return cmd
}
