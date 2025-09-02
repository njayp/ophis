package cfgmgr

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/njayp/ophis/tools"
	"github.com/spf13/cobra"
)

// ValidateExecutable validates that the given path is an executable file.
// It resolves symlinks and checks file permissions.
func ValidateExecutable(executablePath string) (string, error) {
	// Resolve any symlinks to get the actual path
	resolvedPath, err := filepath.EvalSymlinks(executablePath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve executable symlinks at '%s': %w", executablePath, err)
	}

	// Validate that the executable exists and is executable
	stat, err := os.Stat(resolvedPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("executable not found at path '%s': ensure the binary is built and accessible", resolvedPath)
		}
		return "", fmt.Errorf("failed to access executable at '%s': %w", resolvedPath, err)
	}

	if stat.Mode()&0o111 == 0 {
		return "", fmt.Errorf("file at '%s' is not executable: check file permissions", resolvedPath)
	}

	return resolvedPath, nil
}

// DeriveServerName derives a server name from an executable path.
// It uses the base name without extension.
func DeriveServerName(executablePath string) string {
	serverName := filepath.Base(executablePath)
	// Remove extension if present
	if ext := filepath.Ext(serverName); ext != "" {
		serverName = serverName[:len(serverName)-len(ext)]
	}
	return serverName
}

// GetExecutableServerName gets the server name for the current executable.
// If serverName is provided, it returns that. Otherwise, it derives the name
// from the current executable path.
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

// GetMCPCommandPath constructs the command path for invoking the MCP server. It searches for the MCP command up the
// tree and builds the path to execute it from that point up. For example if the command path was
// `<command> alpha mcp start` then this will return `alpha mcp`.
func GetMCPCommandPath(cmd *cobra.Command) []string {
	args := []string{}
	foundMCP := false
	cur := cmd
	for {
		if cur.Name() == tools.MCPCommandName {
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

// BackupConfigFile creates a backup of a configuration file.
// If the file doesn't exist, it returns nil (no error).
// The backup is created with a .backup extension.
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

// LoadJSONConfig loads a JSON configuration file into the provided interface.
// If the file doesn't exist, it returns nil error and leaves the config unchanged.
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

// SaveJSONConfig saves a configuration to a JSON file with proper formatting.
// It ensures the directory exists before writing.
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

// CheckExecutableExists checks if an executable file exists at the given path.
// Returns true if the file exists, false otherwise.
func CheckExecutableExists(executablePath string) bool {
	_, err := os.Stat(executablePath)
	return err == nil
}
