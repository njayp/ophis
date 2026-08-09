package zed

import (
	"github.com/spf13/cobra"
)

// Command creates a new Cobra command for managing Zed MCP context servers.
// commandName is the name of the ophis command in the Cobra tree (e.g. "mcp" or "agent"),
// used by enable to build the correct command path for editor config files.
// defaultEnv is merged into the server env on enable; user --env values take precedence.
func Command(commandName string, defaultEnv map[string]string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "zed",
		Short: "Manage Zed MCP context servers",
		Long:  "Manage MCP context server configuration for Zed",
	}

	// Add subcommands
	cmd.AddCommand(enableCommand(commandName, defaultEnv), disableCommand(), listCommand())
	return cmd
}
