package cursor

import (
	"github.com/spf13/cobra"
)

const (
	// Command names
	cmdCursorUse   = "cursor"
	cmdCursorShort = "Configure Cursor MCP servers"
	cmdCursorLong  = "Configure MCP servers for Cursor"

	cmdEnableShort = "Add server to Cursor config"
	cmdEnableLong  = "Add this application as an MCP server in Cursor"

	cmdDisableShort = "Remove server from Cursor config"
	cmdDisableLong  = "Remove this application from Cursor MCP servers"

	cmdListShort = "Show Cursor MCP servers"
	cmdListLong  = "Show all MCP servers configured in Cursor"
)

// Command creates a new Cobra command that manages Cursor MCP configuration
// This command can be added as a subcommand to any Cobra-based application
func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   cmdCursorUse,
		Short: cmdCursorShort,
		Long:  cmdCursorLong,
	}

	// Add subcommands
	cmd.AddCommand(enableCommand(), disableCommand(), listCommand())
	return cmd
}
