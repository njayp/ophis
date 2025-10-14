package vscode

import (
	"os"
	"path/filepath"
)

// Input represents a VSCode input variable configuration.
type Input struct {
	Type        string `json:"type"`
	ID          string `json:"id"`
	Description string `json:"description"`
	Password    bool   `json:"password,omitempty"`
}

// ConfigPath returns the provided path if non-empty, otherwise returns the
// platform-specific default path for VSCode configuration.
// If workspace is true, returns workspace configuration path (.vscode/mcp.json),
// otherwise returns user-level configuration path.
func ConfigPath(workspace bool) string {
	if workspace {
		return getDefaultWorkspaceConfigPath()
	}
	return getDefaultVSCodeUserConfigPath()
}

// getDefaultWorkspaceConfigPath returns the default workspace configuration path (.vscode/mcp.json).
func getDefaultWorkspaceConfigPath() string {
	workingDir, err := os.Getwd()
	if err != nil {
		// Fallback to current directory
		return filepath.Join(".vscode", "mcp.json")
	}

	return filepath.Join(workingDir, ".vscode", "mcp.json")
}
