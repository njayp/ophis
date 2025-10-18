package claude

import (
	"fmt"
	"os"

	"github.com/njayp/ophis/internal/cfgmgr/manager"
	"github.com/njayp/ophis/internal/cfgmgr/manager/claude"
	"github.com/spf13/cobra"
)

type enableFlags struct {
	configPath string
	logLevel   string
	serverName string
	env        map[string]string
}

// enableCommand creates a Cobra command for adding an MCP server to Claude Desktop.
func enableCommand() *cobra.Command {
	f := &enableFlags{}
	cmd := &cobra.Command{
		Use:   "enable",
		Short: "Add server to Claude config",
		Long:  "Add this application as an MCP server in Claude Desktop",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return f.run(cmd)
		},
	}

	// Add flags
	flags := cmd.Flags()
	flags.StringVar(&f.logLevel, "log-level", "", "Log level (debug, info, warn, error)")
	flags.StringVar(&f.configPath, "config-path", "", "Path to Claude config file")
	flags.StringVar(&f.serverName, "server-name", "", "Name for the MCP server (default: derived from executable name)")
	flags.StringToStringVarP(&f.env, "env", "e", nil, "Environment variables (e.g., --env KEY1=value1 --env KEY2=value2)")
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

	server := claude.MCPServer{
		Command: executablePath,
		Args:    append(mcpPath, "start"),
	}

	// Add log level to args if specified
	if f.logLevel != "" {
		server.Args = append(server.Args, "--log-level", f.logLevel)
	}

	// Add environment variables if specified
	if len(f.env) > 0 {
		server.Env = f.env
	}

	if f.serverName == "" {
		f.serverName = manager.DeriveServerName(executablePath)
	}

	m, err := manager.NewClaudeManager(f.configPath)
	if err != nil {
		return err
	}

	return m.EnableServer(f.serverName, server)
}
