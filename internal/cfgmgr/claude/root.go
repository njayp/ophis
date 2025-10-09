package claude

import (
	"github.com/spf13/cobra"
)

const (
	// Command names
	cmdClaudeUse   = "claude"
	cmdClaudeShort = "Configure Claude Desktop MCP servers"
	cmdClaudeLong  = "Configure MCP servers for Claude Desktop"

	cmdEnableShort = "Add server to Claude config"
	cmdEnableLong  = "Add this application as an MCP server in Claude Desktop"

	cmdDisableShort = "Remove server from Claude config"
	cmdDisableLong  = "Remove this application from Claude Desktop MCP servers"

	cmdListShort = "Show Claude MCP servers"
	cmdListLong  = "Show all MCP servers configured in Claude Desktop"
)

// Command creates a new Cobra command that manages Claude MCP configuration
// This command can be added as a subcommand to any Cobra-based application
func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   cmdClaudeUse,
		Short: cmdClaudeShort,
		Long:  cmdClaudeLong,
	}

	// Add subcommands
	cmd.AddCommand(enableCommand(), disableCommand(), listCommand())
	return cmd
}
