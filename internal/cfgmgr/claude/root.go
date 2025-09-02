package claude

import (
	"github.com/spf13/cobra"
)

// Command creates a new Cobra command that manages Claude MCP configuration
// This command can be added as a subcommand to any Cobra-based application
func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claude",
		Short: "Configure Claude Desktop MCP servers",
		Long:  `Configure MCP servers for Claude Desktop`,
	}

	// Add subcommands
	cmd.AddCommand(enableCommand(), disableCommand(), listCommand())
	return cmd
}
