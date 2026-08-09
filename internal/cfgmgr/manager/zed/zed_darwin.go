package zed

import (
	"os"
	"path/filepath"
)

// getDefaultZedUserConfigPath returns the default Zed user settings.json path on macOS
func getDefaultZedUserConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to a reasonable default
		return filepath.Join("/Users", os.Getenv("USER"), ".config", "zed", "settings.json")
	}
	return filepath.Join(homeDir, ".config", "zed", "settings.json")
}
