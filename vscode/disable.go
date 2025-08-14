package vscode

import (
	"fmt"

	"github.com/njayp/ophis/internal/mcpconfig"
	"github.com/njayp/ophis/vscode/config"
	"github.com/spf13/cobra"
)

type disableCommandFlags struct {
	configPath string
	serverName string
	workspace  bool
	configType string
}

// disableCommand creates a Cobra command for disabling the MCP server in VSCode.
func disableCommand() *cobra.Command {
	disableFlags := &disableCommandFlags{}
	cmd := &cobra.Command{
		Use:   "disable",
		Short: "Disable the MCP server in VSCode",
		Long:  `Disable the MCP server by removing it from VSCode's MCP configuration file (.vscode/mcp.json or user mcp.json)`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return disableMCPServer(disableFlags)
		},
	}

	// Add flags
	flags := cmd.Flags()
	flags.StringVar(&disableFlags.configPath, "config-path", "", "Path to VSCode config file")
	flags.StringVar(&disableFlags.serverName, "server-name", "", "Name of the MCP server to remove (default: derived from executable name)")
	flags.BoolVar(&disableFlags.workspace, "workspace", false, "Remove from workspace settings (.vscode/mcp.json) instead of user settings")
	flags.StringVar(&disableFlags.configType, "config-type", "", "Configuration type: 'workspace' or 'user' (default: user)")

	return cmd
}

func disableMCPServer(flags *disableCommandFlags) error {
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
	serverName, err := mcpconfig.GetExecutableServerName(flags.serverName)
	if err != nil {
		return err
	}

	// Check if server exists
	exists, err := configManager.HasServer(serverName)
	if err != nil {
		return fmt.Errorf("failed to check if MCP server '%s' exists in VSCode configuration: %w", serverName, err)
	}
	if !exists {
		fmt.Printf("MCP server '%s' is not currently enabled in VSCode\n", serverName)
		return nil
	}

	// Remove server from config (with backup)
	if err := configManager.BackupConfig(); err != nil {
		fmt.Printf("Warning: failed to create backup: %v\n", err)
	}

	if err := configManager.RemoveServer(serverName); err != nil {
		return fmt.Errorf("failed to remove MCP server '%s' from VSCode configuration: %w", serverName, err)
	}

	configTypeStr := "user"
	if configType == config.WorkspaceConfig {
		configTypeStr = "workspace"
	}

	fmt.Printf("Successfully disabled MCP server '%s' from VSCode (%s configuration)\n", serverName, configTypeStr)
	fmt.Printf("Configuration file: %s\n", configManager.GetConfigPath())
	fmt.Printf("\nTo apply changes, restart VSCode or reload the window.\n")
	return nil
}
