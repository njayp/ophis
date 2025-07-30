package ophis

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/njayp/ophis/bridge"
	"github.com/spf13/cobra"
)

// toolCommand creates a command that outputs available tools to a file
func toolCommand(config *bridge.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tools",
		Short: "Export available MCP tools to JSON file",
		Long:  `Export all available MCP tools to mcp-tools.json for inspection and debugging.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if config == nil {
				config = &bridge.Config{}
			}
			validateConfig(config, cmd)

			tools := config.Tools()
			mcpTools := make([]mcp.Tool, len(tools))
			for i, tool := range tools {
				mcpTools[i] = tool.Tool
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

			cmd.Printf("Successfully exported %d tools to mcp-tools.json\n", len(tools))
			return nil
		},
	}

	return cmd
}
