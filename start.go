package ophis

import (
	"log/slog"
	"strings"

	"github.com/njayp/ophis/internal/bridge"
	"github.com/njayp/ophis/tools"
	"github.com/spf13/cobra"
)

// StartCommandFlags holds configuration flags for the start command.
type StartCommandFlags struct {
	LogLevel string
}

// startCommand creates a Cobra command for starting the MCP server.
func startCommand(config *Config) *cobra.Command {
	mcpFlags := &StartCommandFlags{}
	cmd := &cobra.Command{
		Use:   tools.StartCommandName,
		Short: "Start the MCP server",
		Long:  `Start MCP server to expose CLI commands to AI assistants`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if config == nil {
				config = &Config{}
			}

			if mcpFlags.LogLevel != "" {
				level := parseLogLevel(mcpFlags.LogLevel)
				// Ensure SloggerOptions is initialized
				if config.SloggerOptions == nil {
					config.SloggerOptions = &slog.HandlerOptions{}
				}
				// Set the log level based on the flag
				config.SloggerOptions.Level = level
			}

			rootCmd := cmd.Parent().Parent()
			if config.RootCmd != nil {
				rootCmd = config.RootCmd
			}

			// Create and start the bridge
			return bridge.Run(config.bridgeConfig(rootCmd))
		},
	}

	// Add flags
	flags := cmd.Flags()
	flags.StringVar(&mcpFlags.LogLevel, "log-level", "", "Log level (debug, info, warn, error)")
	return cmd
}

// parseLogLevel converts a string log level to slog.Level.
// Supported levels are: debug, info, warn, error (case-insensitive).
// Defaults to info for unknown levels.
func parseLogLevel(level string) slog.Level {
	// Parse log level
	slogLevel := slog.LevelInfo
	switch strings.ToLower(level) {
	case "debug":
		slogLevel = slog.LevelDebug
	case "info":
		slogLevel = slog.LevelInfo
	case "warn":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	}

	return slogLevel
}
