package vscode

import (
	"fmt"

	"github.com/njayp/ophis/internal/cfgmgr"
	"github.com/njayp/ophis/internal/cfgmgr/vscode/config"
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
		Use:   cfgmgr.CmdDisable,
		Short: cmdDisableShort,
		Long:  cmdDisableLong,
		RunE: func(_ *cobra.Command, _ []string) error {
			return disableMCPServer(disableFlags)
		},
	}

	// Add flags
	flags := cmd.Flags()
	flags.StringVar(&disableFlags.configPath, cfgmgr.FlagConfigPath, "", "Path to VSCode config file")
	flags.StringVar(&disableFlags.serverName, cfgmgr.FlagServerName, "", "Name of the MCP server to remove (default: derived from executable name)")
	flags.BoolVar(&disableFlags.workspace, cfgmgr.FlagWorkspace, false, "Remove from workspace settings (.vscode/mcp.json) instead of user settings")

	return cmd
}

func disableMCPServer(flags *disableCommandFlags) error {
	// Determine configuration type
	configType := config.UserConfig
	if flags.workspace {
		configType = config.WorkspaceConfig
	}

	// Create config manager
	configManager := config.NewVSCodeConfigManager(flags.configPath, configType)

	// Determine server name
	serverName, err := cfgmgr.GetExecutableServerName(flags.serverName)
	if err != nil {
		return err
	}

	// Validate server name
	if err := cfgmgr.ValidateServerName(serverName); err != nil {
		return err
	}

	// Check if server exists
	exists, err := configManager.HasServer(serverName)
	if err != nil {
		return fmt.Errorf("failed to check if MCP server %q exists in VSCode configuration: %w", serverName, err)
	}
	if !exists {
		fmt.Printf(cfgmgr.MsgServerNotEnabled, serverName)
		return nil
	}

	// Remove server from config (with backup)
	if err := configManager.BackupConfig(); err != nil {
		return fmt.Errorf("failed to create backup: %v", err)
	}

	if err := configManager.RemoveServer(serverName); err != nil {
		return fmt.Errorf("failed to remove MCP server %q from VSCode configuration: %w", serverName, err)
	}

	fmt.Printf(cfgmgr.MsgServerDisabled, serverName)
	fmt.Printf("Configuration: %s\n", configType)
	fmt.Printf("Configuration file: %s\n", configManager.GetConfigPath())
	fmt.Print(cfgmgr.MsgRestartVSCode)
	return nil
}
