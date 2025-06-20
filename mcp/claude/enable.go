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

type enableCommandFlags struct {
	configPath string
	logLevel   string
	logFile    string
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
	flags.StringVar(&enableFlags.logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	flags.StringVar(&enableFlags.logFile, "log-file", "", "Path to log file (default: user cache)")
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

	// Resolve any symlinks to get the actual path
	executablePath, err = filepath.EvalSymlinks(executablePath)
	if err != nil {
		return fmt.Errorf("failed to resolve executable symlinks at '%s': %w", executablePath, err)
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
	configManager := config.NewClaudeConfigManager(flags.configPath)

	// Determine server name
	serverName := flags.serverName
	if serverName == "" {
		serverName = filepath.Base(executablePath)
		// Remove extension if present
		if ext := filepath.Ext(serverName); ext != "" {
			serverName = serverName[:len(serverName)-len(ext)]
		}
	}

	// Validate server name
	if serverName == "" {
		return fmt.Errorf("MCP server name cannot be empty: unable to derive name from executable path '%s'", executablePath)
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
		Args:    []string{"mcp", "start"},
	}

	// Add log level and log file to args if specified
	if flags.logLevel != "" {
		server.Args = append(server.Args, "--log-level", flags.logLevel)
	}
	if flags.logFile != "" {
		server.Args = append(server.Args, "--log-file", flags.logFile)
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
