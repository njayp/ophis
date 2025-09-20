package vscode

import (
	"fmt"
	"os"

	"github.com/njayp/ophis/internal/cfgmgr"
	"github.com/njayp/ophis/internal/cfgmgr/vscode/config"
	"github.com/spf13/cobra"
)

type enableCommandFlags struct {
	configPath string
	logLevel   string
	serverName string
	workspace  bool
	configType string
}

// enableCommand creates a Cobra command for enabling the MCP server in VSCode.
func enableCommand() *cobra.Command {
	enableFlags := &enableCommandFlags{}
	cmd := &cobra.Command{
		Use:   "enable",
		Short: "Add server to VSCode config",
		Long:  `Add this application as an MCP server in VSCode`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return enableMCPServer(cmd, enableFlags)
		},
	}

	// Add flags
	flags := cmd.Flags()
	flags.StringVar(&enableFlags.logLevel, "log-level", "", "Log level (debug, info, warn, error)")
	flags.StringVar(&enableFlags.configPath, "config-path", "", "Path to VSCode config file")
	flags.StringVar(&enableFlags.serverName, "server-name", "", "Name for the MCP server (default: derived from executable name)")
	flags.BoolVar(&enableFlags.workspace, "workspace", false, "Add to workspace settings (.vscode/mcp.json) instead of user settings")
	flags.StringVar(&enableFlags.configType, "config-type", "", "Configuration type: 'workspace' or 'user' (default: user)")

	return cmd
}

func enableMCPServer(cmd *cobra.Command, flags *enableCommandFlags) error {
	// Get the current executable path
	executablePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path for MCP server registration: %w", err)
	}

	// Determine configuration type
	configType := config.UserConfig
	if flags.workspace || flags.configType == "workspace" {
		configType = config.WorkspaceConfig
	} else if flags.configType == "user" {
		configType = config.UserConfig
	} else if flags.configType != "" {
		return fmt.Errorf("invalid config type %q: must be 'workspace' or 'user'", flags.configType)
	}

	// Create config manager
	configManager := config.NewVSCodeConfigManager(flags.configPath, configType)

	// Determine server name
	serverName := flags.serverName
	if serverName == "" {
		serverName = cfgmgr.DeriveServerName(executablePath)
		if serverName == "" {
			return fmt.Errorf("MCP server name cannot be empty: unable to derive name from executable path %q", executablePath)
		}
	}

	// Check if server already exists
	exists, err := configManager.HasServer(serverName)
	if err != nil {
		return fmt.Errorf("failed to check if MCP server %q exists in VSCode configuration: %w", serverName, err)
	}

	// Build server configuration
	server := config.MCPServer{
		Type:    "stdio",
		Command: executablePath,
		Args:    append(cfgmgr.GetMCPCommandPath(cmd), cfgmgr.StartCommandName),
	}

	// Add log level to args if specified
	if flags.logLevel != "" {
		server.Args = append(server.Args, "--log-level", flags.logLevel)
	}

	// Add server to config (with backup)
	if err := configManager.BackupConfig(); err != nil {
		return fmt.Errorf("failed to create backup: %v", err)
	}

	// Show warning if overwriting existing server
	if exists {
		fmt.Printf("⚠️ MCP server %q already exists and will be overwritten\n", serverName)
	}

	if err := configManager.AddServer(serverName, server); err != nil {
		return fmt.Errorf("failed to add MCP server %q to VSCode configuration: %w", serverName, err)
	}

	configTypeStr := "user"
	if configType == config.WorkspaceConfig {
		configTypeStr = "workspace"
	}

	fmt.Printf("Successfully enabled MCP server %q in VSCode (%s configuration)\n", serverName, configTypeStr)
	fmt.Printf("Executable: %s\n", executablePath)
	fmt.Printf("Args: %v\n", server.Args)
	fmt.Printf("Configuration file: %s\n", configManager.GetConfigPath())
	fmt.Printf("\nTo use this server:\n")
	fmt.Printf("1. Open GitHub Copilot Chat\n")
	fmt.Printf("2. Use agent mode to access MCP tools\n")
	return nil
}
