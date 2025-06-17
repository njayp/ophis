package commands

import (
	"github.com/njayp/ophis/bridge"
	"github.com/njayp/ophis/commands/claude"
	"github.com/spf13/cobra"
)

// MCPCommand creates a new Cobra command that starts an MCP server
// This command can be added as a subcommand to any Cobra-based application
func MCPCommand(factory bridge.CommandFactory, config *bridge.MCPCommandConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use: bridge.MCPCommandName,
	}

	// Add subcommands
	cmd.AddCommand(StartCommand(factory, config), claude.ClaudeCommand())
	return cmd
}
