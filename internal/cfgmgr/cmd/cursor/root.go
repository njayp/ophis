package cursor

import (
	"github.com/spf13/cobra"
)

// Command creates a new Cobra command for managing Cursor MCP servers.
func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cursor",
		Short: "Manage Cursor MCP servers",
		Long:  "Manage MCP server configuration for Cursor",
	}

	// Add subcommands
	cmd.AddCommand(enableCommand(), disableCommand(), listCommand())
	return cmd
}
