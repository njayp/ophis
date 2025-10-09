package ophis

import (
	"github.com/njayp/ophis/internal/cfgmgr"
	"github.com/njayp/ophis/internal/cfgmgr/claude"
	"github.com/njayp/ophis/internal/cfgmgr/cursor"
	"github.com/njayp/ophis/internal/cfgmgr/vscode"
	"github.com/spf13/cobra"
)

// Command creates MCP server management commands for a Cobra CLI.
// Pass nil for default configuration or provide a Config for customization.
func Command(config *Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   cfgmgr.MCPCommandName,
		Short: "MCP server management",
		Long:  `Manage MCP servers for AI assistants and code editors`,
	}

	// Add subcommands
	cmd.AddCommand(startCommand(config), toolCommand(config), claude.Command(), vscode.Command(), cursor.Command())
	return cmd
}
