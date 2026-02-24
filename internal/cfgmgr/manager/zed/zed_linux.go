package zed

import (
	"os"
	"path/filepath"
)

// getDefaultZedUserConfigPath returns the default Zed user settings.json path on Linux
func getDefaultZedUserConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to a reasonable default
		return filepath.Join("/home", os.Getenv("USER"), ".config", "zed", "settings.json")
	}

	// Check for XDG_CONFIG_HOME first
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		configDir = filepath.Join(homeDir, ".config")
	}

	return filepath.Join(configDir, "zed", "settings.json")
}
