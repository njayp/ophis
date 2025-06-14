package ophis

import (
	"fmt"

	"github.com/spf13/cobra"
)

type MCPCommandFlags struct {
	LogLevel string
}

// MCPCommand creates a new Cobra command that starts an MCP server
// This command can be added as a subcommand to any Cobra-based application
func MCPCommand(factory CommandFactory, config *MCPCommandConfig) *cobra.Command {
	mcpFlags := &MCPCommandFlags{}
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Start MCP (Model Context Protocol) server",
		Long: fmt.Sprintf(`Start an MCP server that exposes this application's commands to MCP clients.

The MCP server will expose all available commands as tools that can be called
by AI assistants and other MCP-compatible clients.`),
		RunE: func(cmd *cobra.Command, args []string) error {
			if mcpFlags.LogLevel != "" {
				config.LogLevel = mcpFlags.LogLevel
			}

			// Create and start the bridge
			bridge := NewCobraToMCPBridge(factory, config)
			return bridge.StartServer()
		},
	}

	// Add flags
	flags := cmd.Flags()
	flags.StringVar(&mcpFlags.LogLevel, "log-level", "", "Log level (debug, info, warn, error)")
	return cmd
}
