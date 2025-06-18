package commands

import (
	"encoding/json"
	"os"

	"github.com/njayp/ophis/bridge"
	"github.com/spf13/cobra"
)

// ToolCommand creates a command that outputs available tools to a file
func ToolCommand(factory bridge.CommandFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use: "tools",
		RunE: func(*cobra.Command, []string) error {
			file, err := os.Create("mcp-tools.txt")
			if err != nil {
				return err
			}

			return json.NewEncoder(file).Encode(factory.Tools())
		},
	}

	return cmd
}
