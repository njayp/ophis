package config

import (
	"os"
	"path/filepath"
)

// getDefaultClaudeConfigPath returns the default Claude config path on Linux
func getDefaultClaudeConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to a reasonable default
		return filepath.Join("/home", os.Getenv("USER"), ".config", "Claude", "claude_desktop_config.json")
	}

	// Check for XDG_CONFIG_HOME first
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		configDir = filepath.Join(homeDir, ".config")
	}

	return filepath.Join(configDir, "Claude", "claude_desktop_config.json")
}
