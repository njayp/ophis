package cursor

import (
	"os"
	"path/filepath"
)

// getDefaultCursorUserConfigPath returns the default Cursor user mcp.json path on Windows
func getDefaultCursorUserConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback using USERPROFILE environment variable
		return filepath.Join(os.Getenv("USERPROFILE"), ".cursor", "mcp.json")
	}
	return filepath.Join(homeDir, ".cursor", "mcp.json")
}
