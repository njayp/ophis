package ophis

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/spf13/cobra"
)

// toolCommandFlags holds flags for the tools command.
type toolCommandFlags struct {
	logLevel string
}

// toolCommand creates the 'mcp tools' command to export tool definitions.
func toolCommand(config *Config) *cobra.Command {
	toolFlags := &toolCommandFlags{}
	cmd := &cobra.Command{
		Use:   "tools",
		Short: "Export tools as JSON",
		Long:  `Export available MCP tools to mcp-tools.json for inspection`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if config == nil {
				config = &Config{}
			}

			if toolFlags.logLevel != "" {
				level := parseLogLevel(toolFlags.logLevel)
				// Ensure SloggerOptions is initialized
				if config.SloggerOptions == nil {
					config.SloggerOptions = &slog.HandlerOptions{}
				}
				// Set the log level based on the flag
				config.SloggerOptions.Level = level
			}

			config.setupSlogger()
			controllers := config.tools(getRootCmd(cmd))
			mcpTools := make([]mcp.Tool, len(controllers))
			for i, c := range controllers {
				slog.Debug("MCP tool", "name", c.Tool.Name, "description", c.Tool.Description)
				mcpTools[i] = c.Tool
			}

			file, err := os.OpenFile("mcp-tools.json", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
			if err != nil {
				return fmt.Errorf("failed to create or open mcp-tools.json file: %w", err)
			}
			defer func() {
				if closeErr := file.Close(); closeErr != nil {
					cmd.Printf("Warning: failed to close file: %v\n", closeErr)
				}
			}()

			encoder := json.NewEncoder(file)
			encoder.SetIndent("", "  ")
			err = encoder.Encode(mcpTools)
			if err != nil {
				return fmt.Errorf("failed to encode MCP tools to JSON: %w", err)
			}

			cmd.Printf("Successfully exported %d tools to mcp-tools.json\n", len(controllers))
			return nil
		},
	}

	// Add flags
	flags := cmd.Flags()
	flags.StringVar(&toolFlags.logLevel, "log-level", "", "Log level (debug, info, warn, error)")
	return cmd
}
