package claude

import (
	"github.com/spf13/cobra"
)

// Command creates a new Cobra command for managing Claude Desktop MCP servers.
// commandName is the name of the ophis root command (e.g. "mcp" or "agent"),
// used by enable to build the correct command path for editor config files.
// defaultEnv is merged into the server env on enable; user --env values take precedence.
func Command(commandName string, defaultEnv map[string]string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claude",
		Short: "Manage Claude Desktop MCP servers",
		Long:  "Manage MCP server configuration for Claude Desktop",
	}

	// Add subcommands
	cmd.AddCommand(enableCommand(commandName, defaultEnv), disableCommand(), listCommand())
	return cmd
}
