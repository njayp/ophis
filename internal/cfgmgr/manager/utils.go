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

// ResolveServerName determines the MCP server entry name using the precedence:
// the --server-name flag wins, then the configured default, then a name derived
// from the current executable's file name. Keeping this in one place ensures
// enable and disable resolve the name identically across all editors.
func ResolveServerName(flag, defaultName string) (string, error) {
	if flag != "" {
		return flag, nil
	}
	if defaultName != "" {
		return defaultName, nil
	}
	executablePath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to determine executable path: %w", err)
	}
	return DeriveServerName(executablePath), nil
}

// ServerNameUsage returns the help text for the `--server-name` flag on
// `enable`, describing the default that applies when the flag is omitted.
// defaultServerName is the name configured via Config.ServerName; when empty,
// the name is derived from the executable's file name instead.
func ServerNameUsage(defaultServerName string) string {
	return serverNameUsage("Name for the MCP server", defaultServerName)
}

// ServerNameRemoveUsage returns the help text for the `--server-name` flag on
// `disable`, describing the default that applies when the flag is omitted.
// defaultServerName is the name configured via Config.ServerName; when empty,
// the name is derived from the executable's file name instead.
func ServerNameRemoveUsage(defaultServerName string) string {
	return serverNameUsage("Name of the MCP server to remove", defaultServerName)
}

// serverNameUsage renders the shared default clause for the `--server-name`
// flag onto the given prefix.
func serverNameUsage(prefix, defaultServerName string) string {
	if defaultServerName != "" {
		return fmt.Sprintf("%s (default: %q)", prefix, defaultServerName)
	}
	return fmt.Sprintf("%s (default: derived from executable name)", prefix)
}

// GetCmdPath builds the command path to the ophis command.
// It returns the slice of command names from after the root command up to and
// including the command identified by commandName.
//
// Example: for command path "myapp agent claude enable" with commandName "agent",
// returns ["agent"].
//
// The returned slice can be used as arguments when invoking the executable.
// Returns an error if commandName is not found in the command path.
func GetCmdPath(cmd *cobra.Command, commandName string) ([]string, error) {
	path := cmd.CommandPath()
	args := strings.Fields(path)

	// Find the index of the ophis command in the command path
	index := slices.Index(args, commandName)
	if index == -1 {
		return nil, fmt.Errorf("command %q not found in path %q", commandName, path)
	}

	// Return the slice from after the root command to the ophis command
	return args[1 : index+1], nil
}
