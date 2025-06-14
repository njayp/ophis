package ophis

import (
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/spf13/cobra"
)

type MCPCommandFlags struct {
	LogLevel string
}

// MCPCommand creates a new Cobra command that starts an MCP server
// This command can be added as a subcommand to any Cobra-based application
func MCPCommand(factory CommandFactory, logput io.Writer) *cobra.Command {
	mcpFlags := &MCPCommandFlags{}
	if factory == nil {
		panic("factory cannot be nil")
	}

	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Start MCP (Model Context Protocol) server",
		Long: fmt.Sprintf(`Start an MCP server that exposes this application's commands to MCP clients.

The MCP server will expose all available commands as tools that can be called
by AI assistants and other MCP-compatible clients.`),
		RunE: func(cmd *cobra.Command, args []string) error {
			config := &MCPCommandConfig{}

			// Parse log level
			logLevel := slog.LevelInfo
			switch strings.ToLower(mcpFlags.LogLevel) {
			case "debug":
				logLevel = slog.LevelDebug
			case "info":
				logLevel = slog.LevelInfo
			case "warn":
				logLevel = slog.LevelWarn
			case "error":
				logLevel = slog.LevelError
			}

			// Create logger
			logger := slog.New(slog.NewTextHandler(logput, &slog.HandlerOptions{
				Level: logLevel,
			}))

			logger.Info("Starting MCP server",
				"app_name", config.AppName,
				"app_version", config.AppVersion,
				"log_level", logLevel,
			)

			// Create and start the bridge
			bridge := NewCobraToMCPBridge(factory, config)
			return bridge.StartServer()
		},
	}

	// Add flags
	flags := cmd.Flags()
	flags.StringVar(&mcpFlags.LogLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	return cmd
}
