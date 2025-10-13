package manager

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/cobra"
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
	name := "mcp"
	index := slices.Index(args, name)
	if index == -1 {
		// MCP command not found
		return nil, fmt.Errorf("MCP command name %q not found in command path %q", name, path)
	}

	// Return the slice from after the root command to the MCP command
	return args[1 : index+1], nil
}
