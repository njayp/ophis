package cfgmgr

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

const (
	// MCPCommandName is the command name for MCP functionality.
	MCPCommandName = "mcp"
	// StartCommandName is the subcommand to start the MCP server.
	StartCommandName = "start"
)

// DeriveServerName extracts the server name from an executable path.
func DeriveServerName(executablePath string) string {
	serverName := filepath.Base(executablePath)
	// Remove extension if present
	if ext := filepath.Ext(serverName); ext != "" {
		serverName = serverName[:len(serverName)-len(ext)]
	}
	return serverName
}

// GetExecutableServerName returns the provided name or derives it from the executable.
func GetExecutableServerName(serverName string) (string, error) {
	if serverName != "" {
		return serverName, nil
	}

	// Get the current executable path for default name
	executablePath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path for determining default server name: %w", err)
	}

	derivedName := DeriveServerName(executablePath)
	if derivedName == "" {
		return "", fmt.Errorf("MCP server name cannot be empty: unable to derive name from executable path %q", executablePath)
	}

	return derivedName, nil
}

// GetMCPCommandPath builds the command path to the MCP command.
// Example: for "cli alpha mcp start", returns ["alpha", "mcp"].
func GetMCPCommandPath(cmd *cobra.Command) []string {
	path := cmd.CommandPath()
	args := strings.Fields(path) // splits on spaces, handles multiple spaces

	// Find the index of the MCP command name
	index := slices.Index(args, MCPCommandName)
	if index == -1 {
		// MCP command not found
		panic(fmt.Sprintf("MCP command name %q not found in command path %q", MCPCommandName, path))
	}

	// Return the slice from after the root command to the MCP command
	return args[1 : index+1]
}

// BackupConfigFile creates a .backup copy of the configuration file.
func BackupConfigFile(configPath string) error {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// No config file to backup
		return nil
	}

	backupPath := configPath + ".backup"
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read configuration file for backup at %q: %w", configPath, err)
	}

	if err := os.WriteFile(backupPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write backup configuration file at %q: %w", backupPath, err)
	}

	fmt.Printf("Backup config file created at %q\n", backupPath)
	return nil
}

// LoadJSONConfig unmarshals a JSON file into the provided interface.
func LoadJSONConfig(configPath string, config any) error {
	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// File doesn't exist - caller should handle initialization
		return nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read configuration file at %q: %w", configPath, err)
	}

	if err := json.Unmarshal(data, config); err != nil {
		return fmt.Errorf("failed to parse configuration file at %q: invalid JSON format: %w", configPath, err)
	}

	return nil
}

// SaveJSONConfig marshals and saves configuration as formatted JSON.
func SaveJSONConfig(configPath string, config any) error {
	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		return fmt.Errorf("failed to create configuration directory at %q: %w", filepath.Dir(configPath), err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal configuration to JSON: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write configuration file at %q: %w", configPath, err)
	}

	return nil
}
