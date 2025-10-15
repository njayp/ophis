package ophis

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

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
				if config.SloggerOptions == nil {
					config.SloggerOptions = &slog.HandlerOptions{}
				}

				config.SloggerOptions.Level = parseLogLevel(toolFlags.logLevel)
			}

			tools := config.tools(cmd)

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
			err = encoder.Encode(tools)
			if err != nil {
				return fmt.Errorf("failed to encode MCP tools to JSON: %w", err)
			}

			cmd.Printf("Successfully exported %d tools to mcp-tools.json\n", len(tools))
			return nil
		},
	}

	// Add flags
	flags := cmd.Flags()
	flags.StringVar(&toolFlags.logLevel, "log-level", "", "Log level (debug, info, warn, error)")
	return cmd
}

// promptsCommand creates the 'mcp prompts' command to export prompt definitions.
func promptsCommand(config *Config) *cobra.Command {
	toolFlags := &toolCommandFlags{}
	cmd := &cobra.Command{
		Use:   "prompts",
		Short: "Export prompts as JSON",
		Long:  `Export available MCP prompts to mcp-prompts.json for inspection`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if config == nil {
				config = &Config{}
			}

			if toolFlags.logLevel != "" {
				if config.SloggerOptions == nil {
					config.SloggerOptions = &slog.HandlerOptions{}
				}

				config.SloggerOptions.Level = parseLogLevel(toolFlags.logLevel)
			}

			prompts := config.prompts(cmd)

			file, err := os.OpenFile("mcp-prompts.json", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
			if err != nil {
				return fmt.Errorf("failed to create or open mcp-prompts.json file: %w", err)
			}
			defer func() {
				if closeErr := file.Close(); closeErr != nil {
					cmd.Printf("Warning: failed to close file: %v\n", closeErr)
				}
			}()

			encoder := json.NewEncoder(file)
			encoder.SetIndent("", "  ")
			if err := encoder.Encode(prompts); err != nil {
				return fmt.Errorf("failed to encode MCP prompts to JSON: %w", err)
			}

			cmd.Printf("Successfully exported %d prompts to mcp-prompts.json\n", len(prompts))
			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&toolFlags.logLevel, "log-level", "", "Log level (debug, info, warn, error)")
	return cmd
}

// resourcesCommand creates the 'mcp resources' command to export resource definitions.
func resourcesCommand(config *Config) *cobra.Command {
	toolFlags := &toolCommandFlags{}
	cmd := &cobra.Command{
		Use:   "resources",
		Short: "Export resources as JSON",
		Long:  `Export available MCP resources and templates to mcp-resources.json for inspection`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if config == nil {
				config = &Config{}
			}

			if toolFlags.logLevel != "" {
				if config.SloggerOptions == nil {
					config.SloggerOptions = &slog.HandlerOptions{}
				}
				config.SloggerOptions.Level = parseLogLevel(toolFlags.logLevel)
			}

			type resourcesExport struct {
				Resources         any `json:"resources"`
				ResourceTemplates any `json:"resourceTemplates"`
			}

			export := resourcesExport{
				Resources:         config.resources(cmd),
				ResourceTemplates: config.resourceTemplates(cmd),
			}

			file, err := os.OpenFile("mcp-resources.json", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
			if err != nil {
				return fmt.Errorf("failed to create or open mcp-resources.json file: %w", err)
			}
			defer func() {
				if closeErr := file.Close(); closeErr != nil {
					cmd.Printf("Warning: failed to close file: %v\n", closeErr)
				}
			}()

			encoder := json.NewEncoder(file)
			encoder.SetIndent("", "  ")
			if err := encoder.Encode(export); err != nil {
				return fmt.Errorf("failed to encode MCP resources to JSON: %w", err)
			}

			cmd.Printf("Successfully exported resources to mcp-resources.json\n")
			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&toolFlags.logLevel, "log-level", "", "Log level (debug, info, warn, error)")
	return cmd
}
