package ophis

import (
	"github.com/njayp/ophis/internal/cfgmgr/claude"
	"github.com/njayp/ophis/internal/cfgmgr/vscode"
	"github.com/njayp/ophis/tools"
	"github.com/spf13/cobra"
)

// Command creates a new Cobra command that starts an MCP server
// This command can be added as a subcommand to any Cobra-based application
func Command(config *Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   tools.MCPCommandName,
		Short: "MCP (Modal Context Protocol) Server Tools",
		Long: `MCP (Modal Context Protocol) Server Tools

A programmatically-created mcp server with support for multiple LLMs and IDEs. Use the 'tools' subcommand to export available tools.`,
	}

	// Add subcommands
	cmd.AddCommand(startCommand(config), toolCommand(config), claude.Command(), vscode.Command())
	return cmd
}
