package config

import (
	"os"
	"path/filepath"
)

// getDefaultVSCodeUserConfigPath returns the default VSCode user mcp.json path on macOS
func getDefaultVSCodeUserConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to a reasonable default
		return filepath.Join("/Users", os.Getenv("USER"), "Library", "Application Support", "Code", "User", "mcp.json")
	}
	return filepath.Join(homeDir, "Library", "Application Support", "Code", "User", "mcp.json")
}
