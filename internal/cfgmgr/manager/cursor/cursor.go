package cursor

import (
	"os"
	"path/filepath"
)

// ConfigPath returns the platform-specific default path for Cursor configuration.
// If workspace is true, returns workspace configuration path (.cursor/mcp.json),
// otherwise returns user-level configuration path.
func ConfigPath(workspace bool) string {
	if workspace {
		return getDefaultWorkspaceConfigPath()
	}

	return getDefaultCursorUserConfigPath()
}

// getDefaultWorkspaceConfigPath returns the default workspace configuration path (.vscode/mcp.json).
func getDefaultWorkspaceConfigPath() string {
	workingDir, err := os.Getwd()
	if err != nil {
		// Fallback to current directory
		return filepath.Join(".cursor", "mcp.json")
	}

	return filepath.Join(workingDir, ".cursor", "mcp.json")
}
