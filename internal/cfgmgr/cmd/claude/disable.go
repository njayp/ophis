package claude

import (
	"fmt"
	"os"

	"github.com/njayp/ophis/internal/cfgmgr/manager"
	"github.com/spf13/cobra"
)

type disableFlags struct {
	configPath string
	serverName string
}

// disableCommand creates a Cobra command for removing an MCP server from Claude Desktop.
func disableCommand() *cobra.Command {
	f := &disableFlags{}
	cmd := &cobra.Command{
		Use:   "disable",
		Short: "Remove server from Claude config",
		Long:  "Remove this application from Claude Desktop MCP servers",
		RunE: func(_ *cobra.Command, _ []string) error {
			return f.run()
		},
	}

	// Add flags
	flags := cmd.Flags()
	flags.StringVar(&f.configPath, "config-path", "", "Path to Claude config file")
	flags.StringVar(&f.serverName, "server-name", "", "Name of the MCP server to remove (default: derived from executable name)")
	return cmd
}

func (f *disableFlags) run() error {
	// Get the current executable path
	executablePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to determine executable path: %w", err)
	}

	if f.serverName == "" {
		f.serverName = manager.DeriveServerName(executablePath)
	}

	// Create config m
	m, err := manager.NewClaudeManager(f.configPath)
	if err != nil {
		return err
	}

	return m.DisableServer(f.serverName)
}
