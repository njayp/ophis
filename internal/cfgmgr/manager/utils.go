package manager

import (
	"fmt"
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

// GetCmdPath builds the command path to the MCP command.
// It returns the slice of command names from after the root command up to and including "mcp".
//
// Example: for command path "myapp alpha mcp start", returns ["alpha", "mcp"].
//
// The returned slice can be used as arguments when invoking the executable.
// Returns an error if "mcp" is not found in the command path.
func GetCmdPath(cmd *cobra.Command) ([]string, error) {
	path := cmd.CommandPath()
	args := strings.Fields(path)

	// Find the index of "mcp" in the command path
	name := "mcp"
	index := slices.Index(args, name)
	if index == -1 {
		return nil, fmt.Errorf("command %q not found in path %q", name, path)
	}

	// Return the slice from after the root command to the MCP command
	return args[1 : index+1], nil
}
