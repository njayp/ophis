package vscode

import (
	"fmt"
	"os"

	"github.com/njayp/ophis/internal/cfgmgr"
	"github.com/njayp/ophis/internal/cfgmgr/vscode/config"
	"github.com/njayp/ophis/tools"
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

	// Validate the executable
	executablePath, err = cfgmgr.ValidateExecutable(executablePath)
	if err != nil {
		return err
	}

	// Determine configuration type
	configType := config.UserConfig
	if flags.workspace || flags.configType == "workspace" {
		configType = config.WorkspaceConfig
	} else if flags.configType == "user" {
		configType = config.UserConfig
	} else if flags.configType != "" {
		return fmt.Errorf("invalid config type '%s': must be 'workspace' or 'user'", flags.configType)
	}

	// Create config manager
	configManager := config.NewVSCodeConfigManager(flags.configPath, configType)

	// Determine server name
	serverName := flags.serverName
	if serverName == "" {
		serverName = cfgmgr.DeriveServerName(executablePath)
		if serverName == "" {
			return fmt.Errorf("MCP server name cannot be empty: unable to derive name from executable path '%s'", executablePath)
		}
	}

	// Check if server already exists
	exists, err := configManager.HasServer(serverName)
	if err != nil {
		return fmt.Errorf("failed to check if MCP server '%s' exists in VSCode configuration: %w", serverName, err)
	}
	if exists {
		fmt.Printf("MCP server '%s' is already enabled in VSCode\n", serverName)
		return nil
	}

	// Build server configuration
	server := config.MCPServer{
		Type:    "stdio",
		Command: executablePath,
		Args:    append(cfgmgr.GetMCPCommandPath(cmd), tools.StartCommandName),
	}

	// Add log level to args if specified
	if flags.logLevel != "" {
		server.Args = append(server.Args, "--log-level", flags.logLevel)
	}

	// Add server to config (with backup)
	if err := configManager.BackupConfig(); err != nil {
		fmt.Printf("Warning: failed to create backup: %v\n", err)
	}

	if err := configManager.AddServer(serverName, server); err != nil {
		return fmt.Errorf("failed to add MCP server '%s' to VSCode configuration: %w", serverName, err)
	}

	configTypeStr := "user"
	if configType == config.WorkspaceConfig {
		configTypeStr = "workspace"
	}

	fmt.Printf("Successfully enabled MCP server '%s' in VSCode (%s configuration)\n", serverName, configTypeStr)
	fmt.Printf("Executable: %s\n", executablePath)
	fmt.Printf("Args: %v\n", server.Args)
	fmt.Printf("Configuration file: %s\n", configManager.GetConfigPath())
	fmt.Printf("\nTo use this server:\n")
	fmt.Printf("1. Open GitHub Copilot Chat\n")
	fmt.Printf("2. Use agent mode to access MCP tools\n")
	return nil
}
