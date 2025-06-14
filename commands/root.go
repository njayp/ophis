package commands

import (
	"github.com/ophis/bridge"
	"github.com/spf13/cobra"
)

// MCPCommand creates a new Cobra command that starts an MCP server
// This command can be added as a subcommand to any Cobra-based application
func MCPCommand(factory bridge.CommandFactory, config *bridge.MCPCommandConfig) *cobra.Command {

	cmd := &cobra.Command{
		Use:   bridge.MCPCommandName,
		Short: "Start MCP (Model Context Protocol) server",
		Long: `Start an MCP server that exposes this application's commands to MCP clients.

The MCP server will expose all available commands as tools that can be called
by AI assistants and other MCP-compatible clients.`,
	}

	// Add subcommands
	cmd.AddCommand(StartCommand(factory, config), EnableCommand(), DisableCommand())
	return cmd
}
