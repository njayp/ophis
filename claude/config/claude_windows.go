package config

import (
	"os"
	"path/filepath"
)

// getDefaultClaudeConfigPath returns the default Claude config path on Windows
func getDefaultClaudeConfigPath() string {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		// Fallback to a reasonable default
		userProfile := os.Getenv("USERPROFILE")
		if userProfile != "" {
			appData = filepath.Join(userProfile, "AppData", "Roaming")
		} else {
			appData = "C:\\Users\\Default\\AppData\\Roaming"
		}
	}
	return filepath.Join(appData, "Claude", "claude_desktop_config.json")
}
