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
