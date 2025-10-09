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

// TestVSCodeConfigManager tests the VSCode config manager functionality
func TestVSCodeConfigManager(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "vscode-config-test")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	t.Run("workspace config with mcp.json", func(t *testing.T) {
		configPath := filepath.Join(tempDir, "mcp.json")
		manager := NewVSCodeConfigManager(configPath, WorkspaceConfig)

		// Test loading empty config
		config, err := manager.LoadConfig()
		require.NoError(t, err)
		assert.NotNil(t, config.Servers)
		assert.Len(t, config.Servers, 0)

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
		assert.Len(t, config.Servers, 1)
		assert.Equal(t, server, config.Servers["test-server"])

		// Test removing the server
		err = manager.RemoveServer("test-server")
		require.NoError(t, err)

		exists, err = manager.HasServer("test-server")
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("user config with mcp.json", func(t *testing.T) {
		settingsPath := filepath.Join(tempDir, "mcp.json")
		manager := NewVSCodeConfigManager(settingsPath, UserConfig)

		// Test adding a server to an empty config
		server := MCPServer{
			Type:    cfgmgr.ServerTypeStdio,
			Command: "/path/to/executable",
			Args:    []string{"mcp", "start"},
			Env: map[string]string{
				"DEBUG": "true",
			},
		}

		err = manager.AddServer("test-server", server)
		require.NoError(t, err)

		// Verify the mcp.json was updated correctly
		data, err := os.ReadFile(settingsPath)
		require.NoError(t, err)

		var config Config
		err = json.Unmarshal(data, &config)
		require.NoError(t, err)

		assert.Len(t, config.Servers, 1)
		assert.Equal(t, server, config.Servers["test-server"])

		// Test loading through manager
		loadedConfig, err := manager.LoadConfig()
		require.NoError(t, err)
		assert.Len(t, loadedConfig.Servers, 1)
		assert.Equal(t, server, loadedConfig.Servers["test-server"])
	})

	t.Run("config with inputs", func(t *testing.T) {
		configPath := filepath.Join(tempDir, "mcp-with-inputs.json")
		manager := NewVSCodeConfigManager(configPath, WorkspaceConfig)

		// Create config with inputs
		config := &Config{
			Inputs: []Input{
				{
					Type:        "promptString",
					ID:          "api-key",
					Description: "API Key",
					Password:    true,
				},
			},
			Servers: map[string]MCPServer{
				"test-server": {
					Type:    cfgmgr.ServerTypeStdio,
					Command: "npx",
					Args:    []string{"-y", "@example/server"},
					Env: map[string]string{
						"API_KEY": "${input:api-key}",
					},
				},
			},
		}

		err := manager.SaveConfig(config)
		require.NoError(t, err)

		// Load and verify
		loadedConfig, err := manager.LoadConfig()
		require.NoError(t, err)

		assert.Len(t, loadedConfig.Inputs, 1)
		assert.Equal(t, "promptString", loadedConfig.Inputs[0].Type)
		assert.Equal(t, "api-key", loadedConfig.Inputs[0].ID)
		assert.True(t, loadedConfig.Inputs[0].Password)

		assert.Len(t, loadedConfig.Servers, 1)
		server := loadedConfig.Servers["test-server"]
		assert.Equal(t, "${input:api-key}", server.Env["API_KEY"])
	})

	t.Run("backup functionality", func(t *testing.T) {
		configPath := filepath.Join(tempDir, "backup-test.json")
		manager := NewVSCodeConfigManager(configPath, WorkspaceConfig)

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
		assert.Len(t, config.Servers, 1)
		assert.Equal(t, server, config.Servers["initial-server"])
	})
}

// TestVSCodeConfigManagerEdgeCases tests edge cases and error conditions
func TestVSCodeConfigManagerEdgeCases(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "vscode-config-edge-test")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	t.Run("invalid mcp.json", func(t *testing.T) {
		settingsPath := filepath.Join(tempDir, "invalid-mcp.json")

		// Write invalid JSON
		err := os.WriteFile(settingsPath, []byte("{invalid json"), 0o644)
		require.NoError(t, err)

		manager := NewVSCodeConfigManager(settingsPath, UserConfig)

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

		manager := NewVSCodeConfigManager(settingsPath, UserConfig)

		config, err := manager.LoadConfig()
		require.NoError(t, err)
		assert.NotNil(t, config.Servers)
		assert.Len(t, config.Servers, 0)
	})

	t.Run("nonexistent config file", func(t *testing.T) {
		configPath := filepath.Join(tempDir, "nonexistent.json")
		manager := NewVSCodeConfigManager(configPath, WorkspaceConfig)

		// Should return empty config
		config, err := manager.LoadConfig()
		require.NoError(t, err)
		assert.NotNil(t, config.Servers)
		assert.Len(t, config.Servers, 0)
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
