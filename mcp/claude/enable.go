// Package claude provides Cobra command implementations for MCP server management.
// It includes commands to enable, disable, and list MCP servers in Claude's configuration.
package claude

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/njayp/ophis/mcp/claude/config"
	"github.com/spf13/cobra"
)

// EnableCommandFlags holds configuration flags for enable/disable/list commands.
type EnableCommandFlags struct {
	ConfigPath string
	LogLevel   string
	LogFile    string
	ServerName string
}

// enableCommand creates a Cobra command for enabling the MCP server.
func enableCommand() *cobra.Command {
	enableFlags := &EnableCommandFlags{}
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
	flags.StringVar(&enableFlags.LogLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	flags.StringVar(&enableFlags.LogFile, "log-file", "", "Path to log file (default: user cache)")
	flags.StringVar(&enableFlags.ConfigPath, "config-path", "", "Path to Claude config file")
	flags.StringVar(&enableFlags.ServerName, "server-name", "", "Name for the MCP server (default: derived from executable name)")
	return cmd
}

func enableMCPServer(flags *EnableCommandFlags) error {
	// Get the current executable path
	executablePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Resolve any symlinks to get the actual path
	executablePath, err = filepath.EvalSymlinks(executablePath)
	if err != nil {
		return fmt.Errorf("failed to resolve executable path: %w", err)
	}

	// Validate that the executable exists and is executable
	if stat, err := os.Stat(executablePath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("executable not found at path '%s': ensure the binary is built and accessible", executablePath)
		}
		return fmt.Errorf("failed to access executable at '%s': %w", executablePath, err)
	} else if stat.Mode()&0o111 == 0 {
		return fmt.Errorf("file at '%s' is not executable: check file permissions", executablePath)
	}

	// Create config manager
	configManager := config.NewClaudeConfigManager(flags.ConfigPath)

	// Determine server name
	serverName := flags.ServerName
	if serverName == "" {
		serverName = filepath.Base(executablePath)
		// Remove extension if present
		if ext := filepath.Ext(serverName); ext != "" {
			serverName = serverName[:len(serverName)-len(ext)]
		}
	}

	// Validate server name
	if serverName == "" {
		return fmt.Errorf("server name cannot be empty")
	}

	// Check if server already exists
	exists, err := configManager.HasServer(serverName)
	if err != nil {
		return fmt.Errorf("failed to check if server exists: %w", err)
	}
	if exists {
		fmt.Printf("MCP server '%s' is already enabled\n", serverName)
		return nil
	}

	// Build server configuration
	server := config.MCPServer{
		Command: executablePath,
		Args:    []string{"mcp", "start"},
	}

	// Add log level and log file to args if specified
	if flags.LogLevel != "" {
		server.Args = append(server.Args, "--log-level", flags.LogLevel)
	}
	if flags.LogFile != "" {
		server.Args = append(server.Args, "--log-file", flags.LogFile)
	}

	// Add server to config (with backup)
	if err := configManager.BackupConfig(); err != nil {
		fmt.Printf("Warning: failed to create backup: %v\n", err)
	}

	if err := configManager.AddServer(serverName, server); err != nil {
		return fmt.Errorf("failed to add server to config: %w", err)
	}

	fmt.Printf("Successfully enabled MCP server '%s'\n", serverName)
	fmt.Printf("Executable: %s\n", executablePath)
	fmt.Printf("Args: %v\n", server.Args)
	fmt.Printf("\nTo use this server, restart Claude Desktop.\n")
	return nil
}
