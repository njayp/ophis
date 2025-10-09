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

	// CmdEnable is the enable subcommand name.
	CmdEnable = "enable"
	// CmdDisable is the disable subcommand name.
	CmdDisable = "disable"
	// CmdList is the list subcommand name.
	CmdList = "list"

	// FlagLogLevel is the flag name for log level.
	FlagLogLevel = "log-level"
	// FlagConfigPath is the flag name for config path.
	FlagConfigPath = "config-path"
	// FlagServerName is the flag name for server name.
	FlagServerName = "server-name"
	// FlagWorkspace is the flag name for workspace configuration.
	FlagWorkspace = "workspace"

	// BackupFileSuffix is the suffix for backup files.
	BackupFileSuffix = ".backup"

	// ServerTypeStdio is the stdio server type.
	ServerTypeStdio = "stdio"

	// MaxBackups is the maximum number of backup files to keep.
	MaxBackups = 5
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

// GetCmdPath builds the command path to the MCP command, including the MCP command itself.
// It returns the slice of command names from after the root command up to and including the MCP command.
//
// Example: for a command path "myapp alpha mcp start", this returns ["alpha", "mcp"].
// The returned slice can be used as arguments to invoke the MCP command from the executable.
//
// Returns an error if the MCP command is not found in the command path.
func GetCmdPath(cmd *cobra.Command) ([]string, error) {
	path := cmd.CommandPath()
	args := strings.Fields(path) // splits on spaces, handles multiple spaces

	// Find the index of the MCP command name
	index := slices.Index(args, MCPCommandName)
	if index == -1 {
		// MCP command not found
		return nil, fmt.Errorf("MCP command name %q not found in command path %q", MCPCommandName, path)
	}

	// Return the slice from after the root command to the MCP command
	return args[1 : index+1], nil
}

// BackupConfigFile creates a timestamped backup copy of the configuration file.
// It maintains up to MaxBackups backup files and removes older ones.
func BackupConfigFile(configPath string) error {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// No config file to backup
		return nil
	}

	// Rotate existing backups
	if err := rotateBackups(configPath); err != nil {
		return fmt.Errorf("failed to rotate backups: %w", err)
	}

	// Create new backup with .backup suffix (most recent)
	backupPath := configPath + BackupFileSuffix
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

// rotateBackups shifts existing backups and removes old ones.
func rotateBackups(configPath string) error {
	// Remove the oldest backup if we're at the limit
	oldestBackup := fmt.Sprintf("%s%s.%d", configPath, BackupFileSuffix, MaxBackups-1)
	if _, err := os.Stat(oldestBackup); err == nil {
		if err := os.Remove(oldestBackup); err != nil {
			return fmt.Errorf("failed to remove oldest backup at %q: %w", oldestBackup, err)
		}
	}

	// Shift existing numbered backups
	for i := MaxBackups - 2; i >= 1; i-- {
		oldPath := fmt.Sprintf("%s%s.%d", configPath, BackupFileSuffix, i)
		newPath := fmt.Sprintf("%s%s.%d", configPath, BackupFileSuffix, i+1)
		if _, err := os.Stat(oldPath); err == nil {
			if err := os.Rename(oldPath, newPath); err != nil {
				return fmt.Errorf("failed to rotate backup from %q to %q: %w", oldPath, newPath, err)
			}
		}
	}

	// Move current .backup to .backup.1
	backupPath := configPath + BackupFileSuffix
	if _, err := os.Stat(backupPath); err == nil {
		newPath := backupPath + ".1"
		if err := os.Rename(backupPath, newPath); err != nil {
			return fmt.Errorf("failed to rotate current backup to %q: %w", newPath, err)
		}
	}

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
