package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ophis/config"
	"github.com/spf13/cobra"
)

type EnableCommandFlags struct {
	ConfigPath string
	LogLevel   string
	LogFile    string
	ServerName string
}

func EnableCommand() *cobra.Command {
	enableFlags := &EnableCommandFlags{}
	cmd := &cobra.Command{
		Use:   "enable",
		Short: "Enable the MCP server",
		Long:  `Enable the MCP server by adding it to Claude's MCP config file`,
		RunE: func(cmd *cobra.Command, args []string) error {
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

func DisableCommand() *cobra.Command {
	disableFlags := &EnableCommandFlags{} // Reuse flags struct
	cmd := &cobra.Command{
		Use:   "disable",
		Short: "Disable the MCP server",
		Long:  `Disable the MCP server by removing it from Claude's MCP config file`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return disableMCPServer(disableFlags)
		},
	}

	// Add flags
	flags := cmd.Flags()
	flags.StringVar(&disableFlags.ConfigPath, "config-path", "", "Path to Claude config file")
	flags.StringVar(&disableFlags.ServerName, "server-name", "", "Name of the MCP server to remove (default: derived from executable name)")
	return cmd
}

func ListCommand() *cobra.Command {
	listFlags := &EnableCommandFlags{} // Reuse flags struct for config-path
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List configured MCP servers",
		Long:  `List all MCP servers currently configured in Claude's MCP config file`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return listMCPServers(listFlags)
		},
	}

	// Add flags
	flags := cmd.Flags()
	flags.StringVar(&listFlags.ConfigPath, "config-path", "", "Path to Claude config file")
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
		return fmt.Errorf("executable not found: %w", err)
	} else if stat.Mode()&0o111 == 0 {
		return fmt.Errorf("file is not executable: %s", executablePath)
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

func disableMCPServer(flags *EnableCommandFlags) error {
	// Create config manager
	configManager := config.NewClaudeConfigManager(flags.ConfigPath)

	// Determine server name
	serverName := flags.ServerName
	if serverName == "" {
		// Get the current executable path for default name
		executablePath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("failed to get executable path: %w", err)
		}
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

	// Check if server exists
	exists, err := configManager.HasServer(serverName)
	if err != nil {
		return fmt.Errorf("failed to check if server exists: %w", err)
	}
	if !exists {
		fmt.Printf("MCP server '%s' is not currently enabled\n", serverName)
		return nil
	}

	// Remove server from config (with backup)
	if err := configManager.BackupConfig(); err != nil {
		fmt.Printf("Warning: failed to create backup: %v\n", err)
	}

	if err := configManager.RemoveServer(serverName); err != nil {
		return fmt.Errorf("failed to remove server from config: %w", err)
	}

	fmt.Printf("Successfully disabled MCP server '%s'\n", serverName)
	fmt.Printf("\nTo apply changes, restart Claude Desktop.\n")
	return nil
}

func listMCPServers(flags *EnableCommandFlags) error {
	// Create config manager
	configManager := config.NewClaudeConfigManager(flags.ConfigPath)

	// Load config
	claudeConfig, err := configManager.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Printf("Claude MCP Configuration File: %s\n\n", configManager.GetConfigPath())

	if len(claudeConfig.MCPServers) == 0 {
		fmt.Println("No MCP servers are currently configured.")
		fmt.Println("\nTo enable this application as an MCP server, run:")
		fmt.Println("  <your-app> mcp enable")
		return nil
	}

	fmt.Printf("Configured MCP servers (%d):\n\n", len(claudeConfig.MCPServers))
	for name, server := range claudeConfig.MCPServers {
		fmt.Printf("  📦 %s\n", name)
		fmt.Printf("     Command: %s\n", server.Command)
		if len(server.Args) > 0 {
			fmt.Printf("     Args: %v\n", server.Args)
		}
		if len(server.Env) > 0 {
			fmt.Printf("     Environment: %v\n", server.Env)
		}

		// Check if the executable still exists
		if _, err := os.Stat(server.Command); os.IsNotExist(err) {
			fmt.Printf("     ⚠️  Warning: Executable not found\n")
		}
		fmt.Println()
	}

	fmt.Println("💡 Remember to restart Claude Desktop after making changes to the configuration.")
	return nil
}
