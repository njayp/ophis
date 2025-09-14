package cfgmgr

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"

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
		return "", fmt.Errorf("MCP server name cannot be empty: unable to derive name from executable path '%s'", executablePath)
	}

	return derivedName, nil
}

// GetMCPCommandPath builds the command path to the MCP command.
// Example: for "cli alpha mcp start", returns ["alpha", "mcp"].
func GetMCPCommandPath(cmd *cobra.Command) []string {
	args := []string{}
	foundMCP := false
	cur := cmd
	for {
		if cur.Name() == MCPCommandName {
			foundMCP = true
		}
		if foundMCP {
			args = append(args, cur.Name())
		}
		if cur.Parent() == nil {
			break
		}
		cur = cur.Parent()
	}

	if len(args) == 0 {
		return []string{}
	}

	slices.Reverse(args)

	return args[1:]
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
		return fmt.Errorf("failed to read configuration file for backup at '%s': %w", configPath, err)
	}

	if err := os.WriteFile(backupPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write backup configuration file at '%s': %w", backupPath, err)
	}

	return nil
}

// LoadJSONConfig unmarshals a JSON file into the provided interface.
func LoadJSONConfig(configPath string, config interface{}) error {
	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// File doesn't exist - caller should handle initialization
		return nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read configuration file at '%s': %w", configPath, err)
	}

	if err := json.Unmarshal(data, config); err != nil {
		return fmt.Errorf("failed to parse configuration file at '%s': invalid JSON format: %w", configPath, err)
	}

	return nil
}

// SaveJSONConfig marshals and saves configuration as formatted JSON.
func SaveJSONConfig(configPath string, config interface{}) error {
	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		return fmt.Errorf("failed to create configuration directory at '%s': %w", filepath.Dir(configPath), err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal configuration to JSON: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write configuration file at '%s': %w", configPath, err)
	}

	return nil
}
