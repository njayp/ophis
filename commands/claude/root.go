package claude

import (
	"github.com/spf13/cobra"
)

// ClaudeCommand creates a new Cobra command that starts an MCP server
// This command can be added as a subcommand to any Cobra-based application
func ClaudeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "claude",
	}

	// Add subcommands
	cmd.AddCommand(EnableCommand(), DisableCommand(), ListCommand())
	return cmd
}
