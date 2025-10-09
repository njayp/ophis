package config

import (
	"os"
	"path/filepath"
)

// getDefaultCursorUserConfigPath returns the default Cursor user mcp.json path on macOS
func getDefaultCursorUserConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to a reasonable default
		return filepath.Join("/Users", os.Getenv("USER"), ".cursor", "mcp.json")
	}
	return filepath.Join(homeDir, ".cursor", "mcp.json")
}
