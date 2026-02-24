package manager

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/njayp/ophis/internal/cfgmgr/manager/claude"
	"github.com/njayp/ophis/internal/cfgmgr/manager/cursor"
	"github.com/njayp/ophis/internal/cfgmgr/manager/vscode"
	"github.com/njayp/ophis/internal/cfgmgr/manager/zed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClaudeManager(t *testing.T) {
	// Create a temporary directory for test configs
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "claude_config.json")

	t.Run("NewClaudeManager creates empty config", func(t *testing.T) {
		m, err := NewClaudeManager(configPath)
		require.NoError(t, err)
		assert.NotNil(t, m)
		assert.Equal(t, configPath, m.configPath)
	})

	t.Run("EnableServer adds new server", func(t *testing.T) {
		m, err := NewClaudeManager(configPath)
		require.NoError(t, err)

		server := claude.Server{
			Command: "/usr/local/bin/myapp",
			Args:    []string{"mcp", "start"},
		}

		err = m.EnableServer("test-server", server)
		require.NoError(t, err)

		// Verify the server was added
		assert.True(t, m.config.HasServer("test-server"))

		// Verify the config was saved
		data, err := os.ReadFile(configPath)
		require.NoError(t, err)

		var savedConfig claude.Config
		err = json.Unmarshal(data, &savedConfig)
		require.NoError(t, err)
		assert.True(t, savedConfig.HasServer("test-server"))
	})

	t.Run("EnableServer updates existing server", func(t *testing.T) {
		m, err := NewClaudeManager(configPath)
		require.NoError(t, err)

		originalServer := claude.Server{
			Command: "/usr/local/bin/myapp",
			Args:    []string{"mcp", "start"},
		}

		err = m.EnableServer("test-server", originalServer)
		require.NoError(t, err)

		updatedServer := claude.Server{
			Command: "/usr/local/bin/myapp",
			Args:    []string{"mcp", "start", "--log-level", "debug"},
		}

		err = m.EnableServer("test-server", updatedServer)
		require.NoError(t, err)

		// Reload and verify the update
		m2, err := NewClaudeManager(configPath)
		require.NoError(t, err)
		assert.True(t, m2.config.HasServer("test-server"))
	})

	t.Run("DisableServer removes existing server", func(t *testing.T) {
		m, err := NewClaudeManager(configPath)
		require.NoError(t, err)

		server := claude.Server{
			Command: "/usr/local/bin/myapp",
			Args:    []string{"mcp", "start"},
		}

		err = m.EnableServer("test-server", server)
		require.NoError(t, err)

		err = m.DisableServer("test-server")
		require.NoError(t, err)

		// Verify the server was removed
		assert.False(t, m.config.HasServer("test-server"))

		// Reload and verify persistence
		m2, err := NewClaudeManager(configPath)
		require.NoError(t, err)
		assert.False(t, m2.config.HasServer("test-server"))
	})

	t.Run("DisableServer handles non-existent server", func(t *testing.T) {
		m, err := NewClaudeManager(configPath)
		require.NoError(t, err)

		err = m.DisableServer("non-existent")
		require.NoError(t, err)
	})

	t.Run("loadConfig handles missing file", func(t *testing.T) {
		nonExistentPath := filepath.Join(tmpDir, "non-existent.json")
		m, err := NewClaudeManager(nonExistentPath)
		require.NoError(t, err)
		assert.NotNil(t, m)
	})

	t.Run("loadConfig handles invalid JSON", func(t *testing.T) {
		invalidPath := filepath.Join(tmpDir, "invalid.json")
		err := os.WriteFile(invalidPath, []byte("not valid json"), 0o644)
		require.NoError(t, err)

		_, err = NewClaudeManager(invalidPath)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid JSON format")
	})

	t.Run("backupConfig creates backup", func(t *testing.T) {
		backupTestPath := filepath.Join(tmpDir, "backup_test.json")
		m, err := NewClaudeManager(backupTestPath)
		require.NoError(t, err)

		server := claude.Server{
			Command: "/usr/local/bin/myapp",
			Args:    []string{"mcp", "start"},
		}

		err = m.EnableServer("backup-test", server)
		require.NoError(t, err)

		// Make a change to trigger backup
		err = m.EnableServer("backup-test-2", server)
		require.NoError(t, err)

		// Verify backup exists
		backupPath := filepath.Join(tmpDir, "backup_test.backup.json")
		_, err = os.Stat(backupPath)
		require.NoError(t, err)
	})
}

func TestCursorManager(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "cursor_config.json")

	t.Run("NewCursorManager creates empty config", func(t *testing.T) {
		m, err := NewCursorManager(configPath, false)
		require.NoError(t, err)
		assert.NotNil(t, m)
		assert.Equal(t, configPath, m.configPath)
	})

	t.Run("EnableServer adds new server", func(t *testing.T) {
		m, err := NewCursorManager(configPath, false)
		require.NoError(t, err)

		server := cursor.Server{
			Type:    "stdio",
			Command: "/usr/local/bin/myapp",
			Args:    []string{"mcp", "start"},
		}

		err = m.EnableServer("test-server", server)
		require.NoError(t, err)

		assert.True(t, m.config.HasServer("test-server"))

		data, err := os.ReadFile(configPath)
		require.NoError(t, err)

		var savedConfig cursor.Config
		err = json.Unmarshal(data, &savedConfig)
		require.NoError(t, err)
		assert.True(t, savedConfig.HasServer("test-server"))
	})

	t.Run("EnableServer with environment variables", func(t *testing.T) {
		m, err := NewCursorManager(configPath, false)
		require.NoError(t, err)

		server := cursor.Server{
			Type:    "stdio",
			Command: "/usr/local/bin/myapp",
			Args:    []string{"mcp", "start"},
			Env: map[string]string{
				"DEBUG": "true",
				"PORT":  "9090",
			},
		}

		err = m.EnableServer("env-test", server)
		require.NoError(t, err)

		m2, err := NewCursorManager(configPath, false)
		require.NoError(t, err)
		assert.True(t, m2.config.HasServer("env-test"))
	})

	t.Run("DisableServer removes existing server", func(t *testing.T) {
		m, err := NewCursorManager(configPath, false)
		require.NoError(t, err)

		server := cursor.Server{
			Type:    "stdio",
			Command: "/usr/local/bin/myapp",
			Args:    []string{"mcp", "start"},
		}

		err = m.EnableServer("test-server", server)
		require.NoError(t, err)

		err = m.DisableServer("test-server")
		require.NoError(t, err)

		assert.False(t, m.config.HasServer("test-server"))

		m2, err := NewCursorManager(configPath, false)
		require.NoError(t, err)
		assert.False(t, m2.config.HasServer("test-server"))
	})

	t.Run("DisableServer handles non-existent server", func(t *testing.T) {
		m, err := NewCursorManager(configPath, false)
		require.NoError(t, err)

		err = m.DisableServer("non-existent")
		require.NoError(t, err)
	})

	t.Run("loadConfig handles missing file", func(t *testing.T) {
		nonExistentPath := filepath.Join(tmpDir, "cursor_non_existent.json")
		m, err := NewCursorManager(nonExistentPath, false)
		require.NoError(t, err)
		assert.NotNil(t, m)
	})

	t.Run("loadConfig handles invalid JSON", func(t *testing.T) {
		invalidPath := filepath.Join(tmpDir, "cursor_invalid.json")
		err := os.WriteFile(invalidPath, []byte("not valid json"), 0o644)
		require.NoError(t, err)

		_, err = NewCursorManager(invalidPath, false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid JSON format")
	})

	t.Run("Multiple servers can coexist", func(t *testing.T) {
		m, err := NewCursorManager(configPath, false)
		require.NoError(t, err)

		for _, name := range []string{"server-a", "server-b", "server-c"} {
			err = m.EnableServer(name, cursor.Server{
				Type:    "stdio",
				Command: "/usr/local/bin/" + name,
				Args:    []string{"mcp", "start"},
			})
			require.NoError(t, err)
		}

		assert.True(t, m.config.HasServer("server-a"))
		assert.True(t, m.config.HasServer("server-b"))
		assert.True(t, m.config.HasServer("server-c"))

		err = m.DisableServer("server-b")
		require.NoError(t, err)

		assert.True(t, m.config.HasServer("server-a"))
		assert.False(t, m.config.HasServer("server-b"))
		assert.True(t, m.config.HasServer("server-c"))
	})
}

func TestZedManager(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "zed_settings.json")

	t.Run("NewZedManager creates empty config", func(t *testing.T) {
		m, err := NewZedManager(configPath, false)
		require.NoError(t, err)
		assert.NotNil(t, m)
		assert.Equal(t, configPath, m.configPath)
	})

	t.Run("EnableServer adds new server", func(t *testing.T) {
		m, err := NewZedManager(configPath, false)
		require.NoError(t, err)

		server := zed.Server{
			Command: "/usr/local/bin/myapp",
			Args:    []string{"mcp", "start"},
		}

		err = m.EnableServer("test-server", server)
		require.NoError(t, err)

		assert.True(t, m.config.HasServer("test-server"))

		// Verify persisted as valid strict JSON (no comments/trailing commas)
		data, err := os.ReadFile(configPath)
		require.NoError(t, err)

		var savedConfig zed.Config
		err = json.Unmarshal(data, &savedConfig)
		require.NoError(t, err)
		assert.True(t, savedConfig.HasServer("test-server"))
		assert.Equal(t, "/usr/local/bin/myapp", savedConfig.ContextServers["test-server"].Command)
	})

	t.Run("EnableServer with environment variables", func(t *testing.T) {
		m, err := NewZedManager(configPath, false)
		require.NoError(t, err)

		server := zed.Server{
			Command: "/usr/local/bin/myapp",
			Args:    []string{"mcp", "start"},
			Env: map[string]string{
				"DEBUG": "true",
				"PATH":  "/usr/local/bin:/usr/bin",
			},
		}

		err = m.EnableServer("env-test", server)
		require.NoError(t, err)

		m2, err := NewZedManager(configPath, false)
		require.NoError(t, err)
		assert.True(t, m2.config.HasServer("env-test"))
	})

	t.Run("DisableServer removes existing server", func(t *testing.T) {
		m, err := NewZedManager(configPath, false)
		require.NoError(t, err)

		server := zed.Server{
			Command: "/usr/local/bin/myapp",
			Args:    []string{"mcp", "start"},
		}

		err = m.EnableServer("test-server", server)
		require.NoError(t, err)

		err = m.DisableServer("test-server")
		require.NoError(t, err)

		assert.False(t, m.config.HasServer("test-server"))

		m2, err := NewZedManager(configPath, false)
		require.NoError(t, err)
		assert.False(t, m2.config.HasServer("test-server"))
	})

	t.Run("DisableServer handles non-existent server", func(t *testing.T) {
		m, err := NewZedManager(configPath, false)
		require.NoError(t, err)

		err = m.DisableServer("non-existent")
		require.NoError(t, err)
	})

	t.Run("loadConfig handles missing file", func(t *testing.T) {
		nonExistentPath := filepath.Join(tmpDir, "zed_non_existent.json")
		m, err := NewZedManager(nonExistentPath, false)
		require.NoError(t, err)
		assert.NotNil(t, m)
	})

	t.Run("loadConfig handles JSONC with comments and trailing commas", func(t *testing.T) {
		jsoncPath := filepath.Join(tmpDir, "zed_jsonc.json")
		jsoncContent := []byte(`// Zed settings
{
  "theme": "One Dark",
  "ui_font_size": 16,
  "context_servers": {
    "existing-server": {
      "command": "/usr/local/bin/existing",
      "args": ["mcp", "start"],
    },
  },
}`)
		err := os.WriteFile(jsoncPath, jsoncContent, 0o644)
		require.NoError(t, err)

		m, err := NewZedManager(jsoncPath, false)
		require.NoError(t, err)
		assert.True(t, m.config.HasServer("existing-server"))
	})

	t.Run("EnableServer preserves other Zed settings from JSONC file", func(t *testing.T) {
		jsoncPath := filepath.Join(tmpDir, "zed_preserve.json")
		jsoncContent := []byte(`// Zed settings
{
  "theme": "Gruvbox",
  "ui_font_size": 14,
  "context_servers": {
    "pre-existing": {
      "command": "/usr/local/bin/other",
    },
  },
}`)
		err := os.WriteFile(jsoncPath, jsoncContent, 0o644)
		require.NoError(t, err)

		m, err := NewZedManager(jsoncPath, false)
		require.NoError(t, err)

		err = m.EnableServer("new-server", zed.Server{
			Command: "/usr/local/bin/myapp",
			Args:    []string{"mcp", "start"},
		})
		require.NoError(t, err)

		// Reload and verify both servers and other settings are present
		m2, err := NewZedManager(jsoncPath, false)
		require.NoError(t, err)
		assert.True(t, m2.config.HasServer("pre-existing"))
		assert.True(t, m2.config.HasServer("new-server"))

		// Verify the written file is strict JSON (loadable without JSONC handling)
		data, err := os.ReadFile(jsoncPath)
		require.NoError(t, err)
		var raw map[string]json.RawMessage
		err = json.Unmarshal(data, &raw)
		require.NoError(t, err)
		assert.Contains(t, raw, "theme")
		assert.Contains(t, raw, "ui_font_size")
		assert.Contains(t, raw, "context_servers")
	})

	t.Run("loadConfig handles invalid content", func(t *testing.T) {
		invalidPath := filepath.Join(tmpDir, "zed_invalid.json")
		err := os.WriteFile(invalidPath, []byte(`{ completely: broken ??? `), 0o644)
		require.NoError(t, err)

		_, err = NewZedManager(invalidPath, false)
		require.Error(t, err)
	})

	t.Run("Multiple servers can coexist", func(t *testing.T) {
		m, err := NewZedManager(configPath, false)
		require.NoError(t, err)

		for _, name := range []string{"server-a", "server-b", "server-c"} {
			err = m.EnableServer(name, zed.Server{
				Command: "/usr/local/bin/" + name,
				Args:    []string{"mcp", "start"},
			})
			require.NoError(t, err)
		}

		assert.True(t, m.config.HasServer("server-a"))
		assert.True(t, m.config.HasServer("server-b"))
		assert.True(t, m.config.HasServer("server-c"))

		err = m.DisableServer("server-b")
		require.NoError(t, err)

		assert.True(t, m.config.HasServer("server-a"))
		assert.False(t, m.config.HasServer("server-b"))
		assert.True(t, m.config.HasServer("server-c"))
	})

	t.Run("backupConfig creates backup", func(t *testing.T) {
		backupTestPath := filepath.Join(tmpDir, "zed_backup_test.json")
		m, err := NewZedManager(backupTestPath, false)
		require.NoError(t, err)

		err = m.EnableServer("backup-test", zed.Server{Command: "/usr/local/bin/myapp"})
		require.NoError(t, err)

		// Second write triggers backup
		err = m.EnableServer("backup-test-2", zed.Server{Command: "/usr/local/bin/myapp"})
		require.NoError(t, err)

		backupPath := filepath.Join(tmpDir, "zed_backup_test.backup.json")
		_, err = os.Stat(backupPath)
		require.NoError(t, err)
	})
}

func TestVSCodeManager(t *testing.T) {
	// Create a temporary directory for test configs
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "vscode_config.json")

	t.Run("NewVSCodeManager creates empty config", func(t *testing.T) {
		m, err := NewVSCodeManager(configPath, false)
		require.NoError(t, err)
		assert.NotNil(t, m)
		assert.Equal(t, configPath, m.configPath)
	})

	t.Run("EnableServer adds new server", func(t *testing.T) {
		m, err := NewVSCodeManager(configPath, false)
		require.NoError(t, err)

		server := vscode.Server{
			Type:    "stdio",
			Command: "/usr/local/bin/myapp",
			Args:    []string{"mcp", "start"},
		}

		err = m.EnableServer("test-server", server)
		require.NoError(t, err)

		// Verify the server was added
		assert.True(t, m.config.HasServer("test-server"))

		// Verify the config was saved
		data, err := os.ReadFile(configPath)
		require.NoError(t, err)

		var savedConfig vscode.Config
		err = json.Unmarshal(data, &savedConfig)
		require.NoError(t, err)
		assert.True(t, savedConfig.HasServer("test-server"))
	})

	t.Run("EnableServer with environment variables", func(t *testing.T) {
		m, err := NewVSCodeManager(configPath, false)
		require.NoError(t, err)

		server := vscode.Server{
			Type:    "stdio",
			Command: "/usr/local/bin/myapp",
			Args:    []string{"mcp", "start"},
			Env: map[string]string{
				"DEBUG": "true",
				"PORT":  "8080",
			},
		}

		err = m.EnableServer("env-test", server)
		require.NoError(t, err)

		// Reload and verify
		m2, err := NewVSCodeManager(configPath, false)
		require.NoError(t, err)
		assert.True(t, m2.config.HasServer("env-test"))
	})

	t.Run("DisableServer removes existing server", func(t *testing.T) {
		m, err := NewVSCodeManager(configPath, false)
		require.NoError(t, err)

		server := vscode.Server{
			Type:    "stdio",
			Command: "/usr/local/bin/myapp",
			Args:    []string{"mcp", "start"},
		}

		err = m.EnableServer("test-server", server)
		require.NoError(t, err)

		err = m.DisableServer("test-server")
		require.NoError(t, err)

		// Verify the server was removed
		assert.False(t, m.config.HasServer("test-server"))
	})

	t.Run("Multiple servers can coexist", func(t *testing.T) {
		m, err := NewVSCodeManager(configPath, false)
		require.NoError(t, err)

		server1 := vscode.Server{
			Type:    "stdio",
			Command: "/usr/local/bin/app1",
			Args:    []string{"mcp", "start"},
		}

		server2 := vscode.Server{
			Type:    "stdio",
			Command: "/usr/local/bin/app2",
			Args:    []string{"mcp", "start"},
		}

		err = m.EnableServer("server1", server1)
		require.NoError(t, err)

		err = m.EnableServer("server2", server2)
		require.NoError(t, err)

		assert.True(t, m.config.HasServer("server1"))
		assert.True(t, m.config.HasServer("server2"))

		// Disable one and verify the other remains
		err = m.DisableServer("server1")
		require.NoError(t, err)

		assert.False(t, m.config.HasServer("server1"))
		assert.True(t, m.config.HasServer("server2"))
	})
}
