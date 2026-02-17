package ophis

import (
	"github.com/njayp/ophis/internal/cfgmgr/cmd/claude"
	"github.com/njayp/ophis/internal/cfgmgr/cmd/cursor"
	"github.com/njayp/ophis/internal/cfgmgr/cmd/vscode"
	"github.com/spf13/cobra"
)

// Command creates MCP server management commands for a Cobra CLI.
// Pass nil for default configuration or provide a Config for customization.
func Command(config *Config) *cobra.Command {
	name := config.commandName()
	cmd := &cobra.Command{
		Use:   name,
		Short: "MCP server management",
		Long:  `Manage MCP servers for AI assistants and code editors`,
	}

	// Add subcommands
	cmd.AddCommand(
		startCommand(config),
		toolCommand(config),
		streamCommand(config),
		claude.Command(name),
		vscode.Command(name),
		cursor.Command(name),
	)
	return cmd
}
