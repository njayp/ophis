package claude

import (
	"fmt"

	"github.com/njayp/ophis/internal/cfgmgr"
	"github.com/njayp/ophis/internal/cfgmgr/claude/config"
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
		Use:   cfgmgr.CmdDisable,
		Short: cmdDisableShort,
		Long:  cmdDisableLong,
		RunE: func(_ *cobra.Command, _ []string) error {
			return disableMCPServer(disableFlags)
		},
	}

	// Add flags
	flags := cmd.Flags()
	flags.StringVar(&disableFlags.configPath, cfgmgr.FlagConfigPath, "", "Path to Claude config file")
	flags.StringVar(&disableFlags.serverName, cfgmgr.FlagServerName, "", "Name of the MCP server to remove (default: derived from executable name)")
	return cmd
}

func disableMCPServer(flags *disableCommandFlags) error {
	// Create config manager
	configManager := config.NewClaudeConfigManager(flags.configPath)

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
		return fmt.Errorf("failed to check if MCP server %q exists in Claude configuration: %w", serverName, err)
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
		return fmt.Errorf("failed to remove MCP server %q from Claude configuration: %w", serverName, err)
	}

	fmt.Printf(cfgmgr.MsgServerDisabled, serverName)
	fmt.Print(cfgmgr.MsgRestartClaudeDesktop)
	return nil
}
