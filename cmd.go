package ophis

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

type MCPCommandFlags struct {
	AppName    string
	AppVersion string
	LogLevel   string
	LogFile    string
}

// MCPCommand creates a new Cobra command that starts an MCP server
// This command can be added as a subcommand to any Cobra-based application
func MCPCommand(factory CommandFactory) *cobra.Command {
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
			logger := config.Logger
			if logger == nil {
				logger = createLogger(logLevel, mcpFlags.LogFile)
			}

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
	// TODO set defaults
	flags.StringVar(&mcpFlags.LogFile, "log-file", "", "Path to log file (default: stderr)")
	flags.StringVar(&mcpFlags.LogLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	flags.StringVar(&mcpFlags.AppName, "app-name", "", "Application name for MCP server")
	flags.StringVar(&mcpFlags.AppVersion, "app-version", "", "Application version for MCP server")

	return cmd
}

// createLogger creates a logger with the specified level and output
func createLogger(level slog.Level, logFile string) *slog.Logger {
	var handler slog.Handler

	handlerOptions := &slog.HandlerOptions{
		Level: level,
	}

	if logFile != "" {
		// Try to create/open log file
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			// Fallback to stderr if file creation fails
			fmt.Fprintf(os.Stderr, "Warning: Failed to open log file %s: %v. Using stderr.\n", logFile, err)
			handler = slog.NewTextHandler(os.Stderr, handlerOptions)
		} else {
			handler = slog.NewTextHandler(file, handlerOptions)
		}
	} else {
		// Use stderr by default
		handler = slog.NewTextHandler(os.Stderr, handlerOptions)
	}

	logger := slog.New(handler)
	logger.Info("Logger initialized", "level", level.String())
	return logger
}
