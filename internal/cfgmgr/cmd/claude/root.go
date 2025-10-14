package claude

import (
	"github.com/spf13/cobra"
)

// Command creates a new Cobra command for managing Claude Desktop MCP servers.
func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claude",
		Short: "Manage Claude Desktop MCP servers",
		Long:  "Manage MCP server configuration for Claude Desktop",
	}

	// Add subcommands
	cmd.AddCommand(enableCommand(), disableCommand(), listCommand())
	return cmd
}
