package ophis

import (
	"log/slog"

	"github.com/njayp/ophis/internal/cfgmgr"
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
		Use:   cfgmgr.StartCommandName,
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

			// Create and start the bridge
			return config.bridgeConfig(cmd).ServeIO()
		},
	}

	// Add flags
	flags := cmd.Flags()
	flags.StringVar(&mcpFlags.LogLevel, "log-level", "", "Log level (debug, info, warn, error)")
	return cmd
}
