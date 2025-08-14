// Package mcpconfig provides shared utilities for managing MCP server configurations
// across different platforms (Claude Desktop, VSCode, etc.).
package mcpconfig

import (
	"fmt"
	"os"
	"path/filepath"
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
