// Package commands provides the main MCP command implementations for ophis.
// It includes commands to start MCP servers and manage tools.
package commands

import (
	bridge1 "github.com/njayp/ophis/bridge"
	"github.com/njayp/ophis/bridge/tools"
	"github.com/njayp/ophis/commands/claude"
	"github.com/spf13/cobra"
)

// MCPCommand creates a new Cobra command that starts an MCP server
// This command can be added as a subcommand to any Cobra-based application
func MCPCommand(factory bridge1.CommandFactory, config *bridge1.MCPCommandConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use: tools.MCPCommandName,
	}

	// Add subcommands
	cmd.AddCommand(StartCommand(factory, config), ToolCommand(factory), claude.Command())
	return cmd
}
