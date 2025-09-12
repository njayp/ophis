package ophis

import (
	"github.com/njayp/ophis/internal/cfgmgr"
	"github.com/njayp/ophis/internal/cfgmgr/claude"
	"github.com/njayp/ophis/internal/cfgmgr/vscode"
	"github.com/spf13/cobra"
)

// Command creates a new Cobra command that starts an MCP server
// This command can be added as a subcommand to any Cobra-based application
func Command(config *Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   cfgmgr.MCPCommandName,
		Short: "MCP server management",
		Long:  `Manage MCP servers for AI assistants and code editors`,
	}

	// Add subcommands
	cmd.AddCommand(startCommand(config), streamCommand(config), toolCommand(config), claude.Command(), vscode.Command())
	return cmd
}
