package claude

import (
	"github.com/spf13/cobra"
)

// Command creates a new Cobra command that manages Claude MCP configuration
// This command can be added as a subcommand to any Cobra-based application
func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use: "claude",
	}

	// Add subcommands
	cmd.AddCommand(enableCommand(), disableCommand(), listCommand())
	return cmd
}
