package config

import (
	"os"
	"path/filepath"
)

// getDefaultVSCodeUserConfigPath returns the default VSCode user mcp.json path on Linux
func getDefaultVSCodeUserConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback using USER environment variable
		return filepath.Join("/home", os.Getenv("USER"), ".config", "Code", "User", "mcp.json")
	}
	return filepath.Join(homeDir, ".config", "Code", "User", "mcp.json")
}
