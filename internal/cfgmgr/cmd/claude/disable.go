package claude

import (
	"fmt"
	"os"

	"github.com/njayp/ophis/internal/cfgmgr/manager"
	"github.com/njayp/ophis/internal/cfgmgr/manager/claude"
	"github.com/spf13/cobra"
)

type disableCommandFlags struct {
	configPath string
	serverName string
}

// disableCommand creates a Cobra command for disabling the MCP server.
func disableCommand() *cobra.Command {
	disableFlags := &disableCommandFlags{} // Reuse flags struct
	cmd := &cobra.Command{
		Use:   "disable",
		Short: "Remove server from Claude config",
		Long:  "Remove this application from Claude Desktop MCP servers",
		RunE: func(_ *cobra.Command, _ []string) error {
			return disableFlags.disableMCPServer(disableFlags)
		},
	}

	// Add flags
	flags := cmd.Flags()
	flags.StringVar(&disableFlags.configPath, "config-path", "", "Path to Claude config file")
	flags.StringVar(&disableFlags.serverName, "server-name", "", "Name of the MCP server to remove (default: derived from executable name)")
	return cmd
}

func (f *disableCommandFlags) disableMCPServer(flags *disableCommandFlags) error {
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
	manager := manager.Manager[claude.Config, claude.MCPServer]{
		Platform: claude.NewClaudeCodeConfigManager(),
	}

	return manager.DisableMCPServer(serverName)
}
