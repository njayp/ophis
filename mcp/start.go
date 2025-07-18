package mcp

import (
	"fmt"

	"github.com/njayp/ophis/bridge"
	"github.com/spf13/cobra"
)

// StartCommandFlags holds configuration flags for the start command.
type StartCommandFlags struct {
	LogLevel string
}

// startCommand creates a Cobra command for starting the MCP server.
func startCommand(config *bridge.Config) *cobra.Command {
	mcpFlags := &StartCommandFlags{}
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start MCP (Model Context Protocol) server",
		Long: `Start an MCP server that exposes this application's commands to MCP clients.

The MCP server will expose all available commands as tools that can be called
by AI assistants and other MCP-compatible clients.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if config == nil {
				config = &bridge.Config{}
			}
			validateConfig(config, cmd)

			if mcpFlags.LogLevel != "" {
				config.LogLevel = mcpFlags.LogLevel
			}

			// Create and start the bridge
			bridge, err := bridge.New(config)
			if err != nil {
				return fmt.Errorf("failed to create MCP server bridge: %w", err)
			}
			return bridge.StartServer()
		},
	}

	// Add flags
	flags := cmd.Flags()
	flags.StringVar(&mcpFlags.LogLevel, "log-level", "", "Log level (debug, info, warn, error)")
	return cmd
}
