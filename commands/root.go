package commands

import (
	"github.com/ophis/bridge"
	"github.com/spf13/cobra"
)

// MCPCommand creates a new Cobra command that starts an MCP server
// This command can be added as a subcommand to any Cobra-based application
func MCPCommand(factory bridge.CommandFactory, config *bridge.MCPCommandConfig) *cobra.Command {

	cmd := &cobra.Command{
		Use: bridge.MCPCommandName,
	}

	// Add subcommands
	cmd.AddCommand(StartCommand(factory, config), EnableCommand(), DisableCommand())
	return cmd
}
