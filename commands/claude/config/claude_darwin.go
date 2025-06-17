package config

import (
	"os"
	"path/filepath"
)

// getDefaultClaudeConfigPath returns the default Claude config path on macOS
func getDefaultClaudeConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to a reasonable default
		return filepath.Join("/Users", os.Getenv("USER"), "Library", "Application Support", "Claude", "claude_desktop_config.json")
	}
	return filepath.Join(homeDir, "Library", "Application Support", "Claude", "claude_desktop_config.json")
}
