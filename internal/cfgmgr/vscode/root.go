package vscode

import (
	"github.com/spf13/cobra"
)

const (
	// Command names
	cmdVSCodeUse   = "vscode"
	cmdVSCodeShort = "Configure VSCode MCP servers"
	cmdVSCodeLong  = "Configure MCP servers for Visual Studio Code"

	cmdEnableShort = "Add server to VSCode config"
	cmdEnableLong  = "Add this application as an MCP server in VSCode"

	cmdDisableShort = "Remove server from VSCode config"
	cmdDisableLong  = "Remove this application from VSCode MCP servers"

	cmdListShort = "Show VSCode MCP servers"
	cmdListLong  = "Show all MCP servers configured in VSCode"
)

// Command creates a new Cobra command that manages VSCode MCP configuration
// This command can be added as a subcommand to any Cobra-based application
func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   cmdVSCodeUse,
		Short: cmdVSCodeShort,
		Long:  cmdVSCodeLong,
	}

	// Add subcommands
	cmd.AddCommand(enableCommand(), disableCommand(), listCommand())
	return cmd
}
