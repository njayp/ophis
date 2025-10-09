package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/njayp/ophis/internal/cfgmgr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCursorConfigManager tests the Cursor config manager functionality
func TestCursorConfigManager(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "cursor-config-test")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	t.Run("workspace config with mcp.json", func(t *testing.T) {
		configPath := filepath.Join(tempDir, "mcp.json")
		manager := NewCursorConfigManager(configPath, WorkspaceConfig)

		// Test loading empty config
		config, err := manager.LoadConfig()
		require.NoError(t, err)
		assert.NotNil(t, config.MCPServers)
		assert.Len(t, config.MCPServers, 0)

		// Test adding a server
		server := MCPServer{
			Type:    cfgmgr.ServerTypeStdio,
			Command: "/path/to/executable",
			Args:    []string{"mcp", "start"},
		}

		err = manager.AddServer("test-server", server)
		require.NoError(t, err)

		// Test that the server was added
		exists, err := manager.HasServer("test-server")
		require.NoError(t, err)
		assert.True(t, exists)

		// Test loading the updated config
		config, err = manager.LoadConfig()
		require.NoError(t, err)
		assert.Len(t, config.MCPServers, 1)
		assert.Equal(t, server, config.MCPServers["test-server"])

		// Test removing the server
		err = manager.RemoveServer("test-server")
		require.NoError(t, err)

		exists, err = manager.HasServer("test-server")
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("user config type", func(t *testing.T) {
		configPath := filepath.Join(tempDir, "user-mcp.json")
		manager := NewCursorConfigManager(configPath, UserConfig)

		assert.Equal(t, "user", manager.configType.String())

		server := MCPServer{
			Type:    cfgmgr.ServerTypeStdio,
			Command: "/usr/local/bin/server",
			Args:    []string{"start"},
			Env:     map[string]string{"KEY": "value"},
		}

		err := manager.AddServer("user-server", server)
		require.NoError(t, err)

		config, err := manager.LoadConfig()
		require.NoError(t, err)
		assert.Len(t, config.MCPServers, 1)
		assert.Equal(t, server, config.MCPServers["user-server"])
	})

	t.Run("backup functionality", func(t *testing.T) {
		configPath := filepath.Join(tempDir, "backup-test.json")
		manager := NewCursorConfigManager(configPath, WorkspaceConfig)

		// Create initial config
		server := MCPServer{
			Type:    cfgmgr.ServerTypeStdio,
			Command: "/path/to/executable",
		}
		err := manager.AddServer("initial-server", server)
		require.NoError(t, err)

		// Create backup
		err = manager.BackupConfig()
		require.NoError(t, err)

		// Verify backup exists
		backupPath := configPath + ".backup"
		_, err = os.Stat(backupPath)
		require.NoError(t, err)

		// Verify backup content
		data, err := os.ReadFile(backupPath)
		require.NoError(t, err)

		var config Config
		err = json.Unmarshal(data, &config)
		require.NoError(t, err)
		assert.Len(t, config.MCPServers, 1)
		assert.Equal(t, server, config.MCPServers["initial-server"])
	})
}

// TestCursorConfigManagerEdgeCases tests edge cases and error conditions
func TestCursorConfigManagerEdgeCases(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "cursor-config-edge-test")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	t.Run("invalid mcp.json", func(t *testing.T) {
		settingsPath := filepath.Join(tempDir, "invalid-mcp.json")

		// Write invalid JSON
		err := os.WriteFile(settingsPath, []byte("{invalid json"), 0o644)
		require.NoError(t, err)

		manager := NewCursorConfigManager(settingsPath, UserConfig)

		// Should return an error for invalid JSON
		_, err = manager.LoadConfig()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid JSON format")
	})

	t.Run("empty mcp.json", func(t *testing.T) {
		settingsPath := filepath.Join(tempDir, "empty-mcp.json")

		// Write empty JSON object
		err := os.WriteFile(settingsPath, []byte("{}"), 0o644)
		require.NoError(t, err)

		manager := NewCursorConfigManager(settingsPath, UserConfig)

		config, err := manager.LoadConfig()
		require.NoError(t, err)
		assert.NotNil(t, config.MCPServers)
		assert.Len(t, config.MCPServers, 0)
	})

	t.Run("nonexistent config file", func(t *testing.T) {
		configPath := filepath.Join(tempDir, "nonexistent.json")
		manager := NewCursorConfigManager(configPath, WorkspaceConfig)

		// Should return empty config
		config, err := manager.LoadConfig()
		require.NoError(t, err)
		assert.NotNil(t, config.MCPServers)
		assert.Len(t, config.MCPServers, 0)
	})
}

func TestConfigTypeString(t *testing.T) {
	tests := []struct {
		name     string
		cfgType  Type
		expected string
	}{
		{"WorkspaceConfig", WorkspaceConfig, "workspace"},
		{"UserConfig", UserConfig, "user"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.cfgType.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}
