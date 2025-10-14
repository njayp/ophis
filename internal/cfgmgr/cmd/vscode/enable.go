package vscode

import (
	"fmt"
	"os"

	"github.com/njayp/ophis/internal/cfgmgr/manager"
	"github.com/njayp/ophis/internal/cfgmgr/manager/vscode"
	"github.com/spf13/cobra"
)

type enableFlags struct {
	configPath string
	logLevel   string
	serverName string
	workspace  bool
}

// enableCommand creates a Cobra command for adding an MCP server to VSCode.
func enableCommand() *cobra.Command {
	f := &enableFlags{}
	cmd := &cobra.Command{
		Use:   "enable",
		Short: "Add server to VSCode config",
		Long:  "Add this application as an MCP server in VSCode",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return f.run(cmd)
		},
	}

	// Add flags
	flags := cmd.Flags()
	flags.StringVar(&f.logLevel, "log-level", "", "Log level (debug, info, warn, error)")
	flags.StringVar(&f.configPath, "config-path", "", "Path to VSCode config file")
	flags.StringVar(&f.serverName, "server-name", "", "Name for the MCP server (default: derived from executable name)")
	flags.BoolVar(&f.workspace, "workspace", false, "Add to workspace settings (.vscode/mcp.json) instead of user settings")

	return cmd
}

func (f *enableFlags) run(cmd *cobra.Command) error {
	// Get the current executable path
	executablePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to determine executable path: %w", err)
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

	if f.serverName == "" {
		f.serverName = manager.DeriveServerName(executablePath)
	}

	// Create config m
	m, err := manager.NewVSCodeManager(f.configPath, f.workspace)
	if err != nil {
		return err
	}

	return m.EnableServer(f.serverName, server)
}
