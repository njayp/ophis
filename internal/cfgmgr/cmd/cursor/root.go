package cursor

import (
	"github.com/spf13/cobra"
)

// Command creates a new Cobra command for managing Cursor MCP servers.
// commandName is the Use name of the ophis root command (e.g. "mcp" or "agent"),
// threaded through to enableCommand so that GetCmdPath can locate it.
// serverName is the default MCP server entry name for enable; the --server-name
// flag overrides it, and an empty value falls back to the executable name.
// defaultEnv is merged into the server env on enable; user --env values take precedence.
func Command(commandName, serverName string, defaultEnv map[string]string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cursor",
		Short: "Manage Cursor MCP servers",
		Long:  "Manage MCP server configuration for Cursor",
	}

	// Add subcommands
	cmd.AddCommand(enableCommand(commandName, serverName, defaultEnv), disableCommand(serverName), listCommand())
	return cmd
}
