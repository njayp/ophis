package claude

import (
	"fmt"
	"os"

	"github.com/njayp/ophis/internal/cfgmgr"
	"github.com/njayp/ophis/internal/cfgmgr/claude/config"
	"github.com/spf13/cobra"
)

type enableCommandFlags struct {
	configPath string
	logLevel   string
	serverName string
}

// enableCommand creates a Cobra command for enabling the MCP server.
func enableCommand() *cobra.Command {
	enableFlags := &enableCommandFlags{}
	cmd := &cobra.Command{
		Use:   cfgmgr.CmdEnable,
		Short: cmdEnableShort,
		Long:  cmdEnableLong,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return enableMCPServer(cmd, enableFlags)
		},
	}

	// Add flags
	flags := cmd.Flags()
	flags.StringVar(&enableFlags.logLevel, cfgmgr.FlagLogLevel, "", "Log level (debug, info, warn, error)")
	flags.StringVar(&enableFlags.configPath, cfgmgr.FlagConfigPath, "", "Path to Claude config file")
	flags.StringVar(&enableFlags.serverName, cfgmgr.FlagServerName, "", "Name for the MCP server (default: derived from executable name)")
	return cmd
}

func enableMCPServer(cmd *cobra.Command, flags *enableCommandFlags) error {
	// Get the current executable path
	executablePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path for MCP server registration: %w", err)
	}

	// Create config manager
	configManager := config.NewClaudeConfigManager(flags.configPath)

	// Determine server name
	serverName := flags.serverName
	if serverName == "" {
		serverName = cfgmgr.DeriveServerName(executablePath)
		if serverName == "" {
			return fmt.Errorf("MCP server name cannot be empty: unable to derive name from executable path %q", executablePath)
		}
	}

	// Validate server name
	if err := cfgmgr.ValidateServerName(serverName); err != nil {
		return err
	}

	// Validate log level if provided
	if err := cfgmgr.ValidateLogLevel(flags.logLevel); err != nil {
		return err
	}

	// Check if server already exists
	exists, err := configManager.HasServer(serverName)
	if err != nil {
		return fmt.Errorf("failed to check if MCP server %q exists in Claude configuration: %w", serverName, err)
	}

	// Build server configuration
	mcpPath, err := cfgmgr.GetCmdPath(cmd)
	if err != nil {
		return fmt.Errorf("failed to determine MCP command path: %w", err)
	}

	server := config.MCPServer{
		Command: executablePath,
		Args:    append(mcpPath, cfgmgr.StartCommandName),
	}

	// Add log level to args if specified
	if flags.logLevel != "" {
		server.Args = append(server.Args, "--"+cfgmgr.FlagLogLevel, flags.logLevel)
	}

	// Add server to config (with backup)
	if err := configManager.BackupConfig(); err != nil {
		return fmt.Errorf("failed to create backup: %v", err)
	}

	// Show warning if overwriting existing server
	if exists {
		fmt.Printf(cfgmgr.MsgServerOverwrite, serverName)
	}

	if err := configManager.AddServer(serverName, server); err != nil {
		return fmt.Errorf("failed to add MCP server %q to Claude configuration: %w", serverName, err)
	}

	fmt.Printf(cfgmgr.MsgServerEnabled, serverName)
	fmt.Printf("Executable: %s\n", executablePath)
	fmt.Printf("Args: %v\n", server.Args)
	fmt.Print(cfgmgr.MsgRestartClaudeDesktop)
	return nil
}
