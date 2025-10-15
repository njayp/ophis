package vscode

import (
	"os"
	"path/filepath"
)

// getDefaultVSCodeUserConfigPath returns the default VSCode user mcp.json path on Windows
func getDefaultVSCodeUserConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback using USERPROFILE environment variable
		return filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming", "Code", "User", "mcp.json")
	}
	return filepath.Join(homeDir, "AppData", "Roaming", "Code", "User", "mcp.json")
}
