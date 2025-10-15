package ophis

import (
	"log/slog"

	"github.com/spf13/cobra"
)

// startCommandFlags holds flags for the start command.
type startCommandFlags struct {
	logLevel string
}

// startCommand creates the 'mcp start' command.
func startCommand(config *Config) *cobra.Command {
	mcpFlags := &startCommandFlags{}
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the MCP server",
		Long:  `Start MCP server to expose CLI commands to AI assistants`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if config == nil {
				config = &Config{}
			}

			if mcpFlags.logLevel != "" {
				level := parseLogLevel(mcpFlags.logLevel)
				// Ensure SloggerOptions is initialized
				if config.SloggerOptions == nil {
					config.SloggerOptions = &slog.HandlerOptions{}
				}
				// Set the log level based on the flag
				config.SloggerOptions.Level = level
			}

			// Create and start the bridge
			return config.serveStdio(cmd)
		},
	}

	// Add flags
	flags := cmd.Flags()
	flags.StringVar(&mcpFlags.logLevel, "log-level", "", "Log level (debug, info, warn, error)")
	return cmd
}
