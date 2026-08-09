package zed

import (
	"os"
	"path/filepath"
)

// ConfigPath returns the platform-specific default path for Zed configuration.
// If workspace is true, returns workspace configuration path (.zed/settings.json),
// otherwise returns user-level configuration path.
func ConfigPath(workspace bool) string {
	if workspace {
		return getDefaultWorkspaceConfigPath()
	}

	return getDefaultZedUserConfigPath()
}

// getDefaultWorkspaceConfigPath returns the default workspace configuration path (.zed/settings.json).
func getDefaultWorkspaceConfigPath() string {
	workingDir, err := os.Getwd()
	if err != nil {
		// Fallback to current directory
		return filepath.Join(".zed", "settings.json")
	}

	return filepath.Join(workingDir, ".zed", "settings.json")
}
