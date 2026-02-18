package cursor

import (
	"github.com/spf13/cobra"
)

// Command creates a new Cobra command for managing Cursor MCP servers.
// commandName is the Use name of the ophis root command (e.g. "mcp" or "agent"),
// threaded through to enableCommand so that GetCmdPath can locate it.
func Command(commandName string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cursor",
		Short: "Manage Cursor MCP servers",
		Long:  "Manage MCP server configuration for Cursor",
	}

	// Add subcommands
	cmd.AddCommand(enableCommand(commandName), disableCommand(), listCommand())
	return cmd
}
