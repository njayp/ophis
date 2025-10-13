package vscode

import (
	"fmt"
	"os"

	"github.com/njayp/ophis/internal/cfgmgr/manager"
	"github.com/njayp/ophis/internal/cfgmgr/manager/vscode"
	"github.com/spf13/cobra"
)

type disableCommandFlags struct {
	configPath string
	serverName string
	workspace  bool
}

// disableCommand creates a Cobra command for disabling the MCP server in VSCode.
func disableCommand() *cobra.Command {
	disableFlags := &disableCommandFlags{}
	cmd := &cobra.Command{
		Use:   "disable",
		Short: "Remove server from VSCode config",
		Long:  "Remove this application from VSCode MCP servers",
		RunE: func(_ *cobra.Command, _ []string) error {
			return disableFlags.disableMCPServer()
		},
	}

	// Add flags
	flags := cmd.Flags()
	flags.StringVar(&disableFlags.configPath, "config-path", "", "Path to VSCode config file")
	flags.StringVar(&disableFlags.serverName, "server-name", "", "Name of the MCP server to remove (default: derived from executable name)")
	flags.BoolVar(&disableFlags.workspace, "workspace", false, "Remove from workspace settings (.vscode/mcp.json) instead of user settings")

	return cmd
}

func (f *disableCommandFlags) disableMCPServer() error {
	// Get the current executable path
	executablePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path for MCP server registration: %w", err)
	}

	// Determine server name
	serverName := f.serverName
	if serverName == "" {
		serverName = manager.DeriveServerName(executablePath)
		if serverName == "" {
			return fmt.Errorf("MCP server name cannot be empty: unable to derive name from executable path %q", executablePath)
		}
	}

	// Create config manager
	manager := manager.Manager[vscode.Config, vscode.MCPServer]{
		Platform: vscode.NewVSCodeConfigManager(f.workspace),
	}

	return manager.DisableMCPServer(serverName)
}
