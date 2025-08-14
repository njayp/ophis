// Package ophis provides functionality to convert Cobra CLI applications into MCP servers.
// It includes commands to start MCP servers and manage tools.
package ophis

import (
	"github.com/njayp/ophis/claude"
	"github.com/njayp/ophis/tools"
	"github.com/njayp/ophis/vscode"
	"github.com/spf13/cobra"
)

// Command creates a new Cobra command that starts an MCP server
// This command can be added as a subcommand to any Cobra-based application
func Command(config *Config) *cobra.Command {
	cmd := &cobra.Command{
		Use: tools.MCPCommandName,
	}

	// Add subcommands
	cmd.AddCommand(startCommand(config), toolCommand(config), claude.Command(), vscode.Command())
	return cmd
}
