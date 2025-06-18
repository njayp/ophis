// Package mcp provides the main MCP command implementations for ophis.
// It includes commands to start MCP servers and manage tools.
package mcp

import (
	"github.com/njayp/ophis/bridge"
	"github.com/njayp/ophis/mcp/claude"
	"github.com/njayp/ophis/tools"
	"github.com/spf13/cobra"
)

// Command creates a new Cobra command that starts an MCP server
// This command can be added as a subcommand to any Cobra-based application
func Command(factory bridge.CommandFactory, config *bridge.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use: tools.MCPCommandName,
	}

	// Add subcommands
	cmd.AddCommand(startCommand(factory, config), toolCommand(factory), claude.Command())
	return cmd
}
