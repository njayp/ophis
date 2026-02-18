package vscode

import (
	"github.com/spf13/cobra"
)

// Command creates a new Cobra command for managing VSCode MCP servers.
// commandName is the name of the ophis command in the Cobra tree (e.g. "mcp" or "agent"),
// used by enable to build the correct command path for editor config files.
// defaultEnv is merged into the server env on enable; user --env values take precedence.
func Command(commandName string, defaultEnv map[string]string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vscode",
		Short: "Manage VSCode MCP servers",
		Long:  "Manage MCP server configuration for Visual Studio Code",
	}

	// Add subcommands
	cmd.AddCommand(enableCommand(commandName, defaultEnv), disableCommand(), listCommand())
	return cmd
}
