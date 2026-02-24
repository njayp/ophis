package zed

import (
	"os"
	"path/filepath"
)

// getDefaultZedUserConfigPath returns the default Zed user settings.json path on Windows
func getDefaultZedUserConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback using USERPROFILE environment variable
		return filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming", "Zed", "settings.json")
	}
	return filepath.Join(homeDir, "AppData", "Roaming", "Zed", "settings.json")
}
