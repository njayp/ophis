package vscode

import (
	"fmt"
	"os"

	"github.com/njayp/ophis/internal/cfgmgr/manager"
	"github.com/njayp/ophis/internal/cfgmgr/manager/vscode"
	"github.com/spf13/cobra"
)

type enableCommandFlags struct {
	configPath string
	logLevel   string
	serverName string
	workspace  bool
}

// enableCommand creates a Cobra command for enabling the MCP server in VSCode.
func enableCommand() *cobra.Command {
	enableFlags := &enableCommandFlags{}
	cmd := &cobra.Command{
		Use:   "enable",
		Short: "Add server to VSCode config",
		Long:  "Add this application as an MCP server in VSCode",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return enableFlags.enableMCPServer(cmd)
		},
	}

	// Add flags
	flags := cmd.Flags()
	flags.StringVar(&enableFlags.logLevel, "log-level", "", "Log level (debug, info, warn, error)")
	flags.StringVar(&enableFlags.configPath, "config-path", "", "Path to VSCode config file")
	flags.StringVar(&enableFlags.serverName, "server-name", "", "Name for the MCP server (default: derived from executable name)")
	flags.BoolVar(&enableFlags.workspace, "workspace", false, "Add to workspace settings (.vscode/mcp.json) instead of user settings")

	return cmd
}

func (f *enableCommandFlags) enableMCPServer(cmd *cobra.Command) error {
	// Get the current executable path
	executablePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path for MCP server registration: %w", err)
	}

	// Build server configuration
	mcpPath, err := manager.GetCmdPath(cmd)
	if err != nil {
		return fmt.Errorf("failed to determine MCP command path: %w", err)
	}

	server := vscode.MCPServer{
		Type:    "stdio",
		Command: executablePath,
		Args:    append(mcpPath, "start"),
	}

	// Add log level to args if specified
	if f.logLevel != "" {
		server.Args = append(server.Args, "--log-level", f.logLevel)
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

	return manager.EnableMCPServer(serverName, server)
}
