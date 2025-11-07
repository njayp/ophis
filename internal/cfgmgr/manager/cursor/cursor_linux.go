package cursor

import (
	"os"
	"path/filepath"
)

// getDefaultCursorUserConfigPath returns the default Cursor user mcp.json path on Linux
func getDefaultCursorUserConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback using USER environment variable
		return filepath.Join("/home", os.Getenv("USER"), ".cursor", "mcp.json")
	}
	return filepath.Join(homeDir, ".cursor", "mcp.json")
}
