package claude

import (
	"fmt"
	"os"

	"github.com/njayp/ophis/claude/config"
	"github.com/njayp/ophis/internal/mcpconfig"
	"github.com/njayp/ophis/tools"
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
		Use:   "enable",
		Short: "Enable the MCP server",
		Long:  `Enable the MCP server by adding it to Claude's MCP config file`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return enableMCPServer(enableFlags)
		},
	}

	// Add flags
	flags := cmd.Flags()
	flags.StringVar(&enableFlags.logLevel, "log-level", "", "Log level (debug, info, warn, error)")
	flags.StringVar(&enableFlags.configPath, "config-path", "", "Path to Claude config file")
	flags.StringVar(&enableFlags.serverName, "server-name", "", "Name for the MCP server (default: derived from executable name)")
	return cmd
}

func enableMCPServer(flags *enableCommandFlags) error {
	// Get the current executable path
	executablePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path for MCP server registration: %w", err)
	}

	// Validate the executable
	executablePath, err = mcpconfig.ValidateExecutable(executablePath)
	if err != nil {
		return err
	}

	// Create config manager
	configManager := config.NewClaudeConfigManager(flags.configPath)

	// Determine server name
	serverName := flags.serverName
	if serverName == "" {
		serverName = mcpconfig.DeriveServerName(executablePath)
		if serverName == "" {
			return fmt.Errorf("MCP server name cannot be empty: unable to derive name from executable path '%s'", executablePath)
		}
	}

	// Check if server already exists
	exists, err := configManager.HasServer(serverName)
	if err != nil {
		return fmt.Errorf("failed to check if MCP server '%s' exists in Claude configuration: %w", serverName, err)
	}
	if exists {
		fmt.Printf("MCP server '%s' is already enabled\n", serverName)
		return nil
	}

	// Build server configuration
	server := config.MCPServer{
		Command: executablePath,
		Args:    []string{tools.MCPCommandName, tools.StartCommandName},
	}

	// Add log level and log file to args if specified
	if flags.logLevel != "" {
		server.Args = append(server.Args, "--log-level", flags.logLevel)
	}

	// Add server to config (with backup)
	if err := configManager.BackupConfig(); err != nil {
		fmt.Printf("Warning: failed to create backup: %v\n", err)
	}

	if err := configManager.AddServer(serverName, server); err != nil {
		return fmt.Errorf("failed to add MCP server '%s' to Claude configuration: %w", serverName, err)
	}

	fmt.Printf("Successfully enabled MCP server '%s'\n", serverName)
	fmt.Printf("Executable: %s\n", executablePath)
	fmt.Printf("Args: %v\n", server.Args)
	fmt.Printf("\nTo use this server, restart Claude Desktop.\n")
	return nil
}
