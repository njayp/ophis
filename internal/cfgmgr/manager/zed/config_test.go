package zed

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	t.Run("AddServer adds new server", func(t *testing.T) {
		config := &Config{}
		server := Server{
			Command: "/usr/local/bin/myapp",
			Args:    []string{"mcp", "start"},
		}

		config.AddServer("test", server)

		assert.True(t, config.HasServer("test"))
		assert.Equal(t, 1, len(config.ContextServers))
	})

	t.Run("AddServer initializes map if nil", func(t *testing.T) {
		config := &Config{}
		assert.Nil(t, config.ContextServers)

		server := Server{
			Command: "/usr/local/bin/myapp",
		}

		config.AddServer("test", server)
		assert.NotNil(t, config.ContextServers)
		assert.True(t, config.HasServer("test"))
	})

	t.Run("AddServer updates existing server", func(t *testing.T) {
		config := &Config{}

		originalServer := Server{
			Command: "/usr/local/bin/myapp",
			Args:    []string{"start"},
		}
		config.AddServer("test", originalServer)

		updatedServer := Server{
			Command: "/usr/local/bin/myapp",
			Args:    []string{"start", "--verbose"},
		}
		config.AddServer("test", updatedServer)

		assert.Equal(t, 1, len(config.ContextServers))
		assert.Equal(t, 2, len(config.ContextServers["test"].Args))
	})

	t.Run("HasServer returns false for non-existent server", func(t *testing.T) {
		config := &Config{
			ContextServers: map[string]Server{
				"server1": {Command: "/bin/app"},
			},
		}

		assert.False(t, config.HasServer("non-existent"))
		assert.True(t, config.HasServer("server1"))
	})

	t.Run("RemoveServer removes existing server", func(t *testing.T) {
		config := &Config{
			ContextServers: map[string]Server{
				"server1": {Command: "/bin/app1"},
				"server2": {Command: "/bin/app2"},
			},
		}

		config.RemoveServer("server1")

		assert.False(t, config.HasServer("server1"))
		assert.True(t, config.HasServer("server2"))
		assert.Equal(t, 1, len(config.ContextServers))
	})

	t.Run("RemoveServer handles non-existent server", func(t *testing.T) {
		config := &Config{
			ContextServers: map[string]Server{
				"server1": {Command: "/bin/app"},
			},
		}

		config.RemoveServer("non-existent")

		assert.Equal(t, 1, len(config.ContextServers))
		assert.True(t, config.HasServer("server1"))
	})

	t.Run("Multiple servers can be managed", func(t *testing.T) {
		config := &Config{}

		servers := []struct {
			name   string
			server Server
		}{
			{"kubectl", Server{Command: "/usr/local/bin/kubectl"}},
			{"helm", Server{Command: "/usr/local/bin/helm"}},
			{"argocd", Server{Command: "/usr/local/bin/argocd"}},
		}

		for _, s := range servers {
			config.AddServer(s.name, s.server)
		}

		assert.Equal(t, 3, len(config.ContextServers))
		for _, s := range servers {
			assert.True(t, config.HasServer(s.name))
		}

		config.RemoveServer("helm")
		assert.Equal(t, 2, len(config.ContextServers))
		assert.False(t, config.HasServer("helm"))
	})
}

func TestConfigJSON(t *testing.T) {
	t.Run("UnmarshalJSON extracts context_servers", func(t *testing.T) {
		data := []byte(`{
			"theme": "One Dark",
			"font_size": 14,
			"context_servers": {
				"my-server": {
					"command": "/usr/local/bin/myapp",
					"args": ["mcp", "start"]
				}
			}
		}`)

		config := &Config{}
		err := json.Unmarshal(data, config)
		require.NoError(t, err)

		assert.True(t, config.HasServer("my-server"))
		assert.Equal(t, "/usr/local/bin/myapp", config.ContextServers["my-server"].Command)
		assert.Equal(t, []string{"mcp", "start"}, config.ContextServers["my-server"].Args)
	})

	t.Run("UnmarshalJSON handles missing context_servers", func(t *testing.T) {
		data := []byte(`{
			"theme": "One Dark",
			"font_size": 14
		}`)

		config := &Config{}
		err := json.Unmarshal(data, config)
		require.NoError(t, err)

		assert.Nil(t, config.ContextServers)
		assert.Equal(t, 2, len(config.extra))
	})

	t.Run("MarshalJSON preserves other settings", func(t *testing.T) {
		data := []byte(`{"font_size":14,"theme":"One Dark","context_servers":{"my-server":{"command":"/usr/local/bin/myapp"}}}`)

		config := &Config{}
		err := json.Unmarshal(data, config)
		require.NoError(t, err)

		// Add a new server
		config.AddServer("new-server", Server{Command: "/usr/local/bin/other"})

		out, err := json.Marshal(config)
		require.NoError(t, err)

		// Round-trip back to verify all data is present
		var result map[string]json.RawMessage
		err = json.Unmarshal(out, &result)
		require.NoError(t, err)

		assert.Contains(t, result, "theme")
		assert.Contains(t, result, "font_size")
		assert.Contains(t, result, "context_servers")

		var servers map[string]Server
		err = json.Unmarshal(result["context_servers"], &servers)
		require.NoError(t, err)
		assert.True(t, len(servers) == 2)
		assert.Contains(t, servers, "my-server")
		assert.Contains(t, servers, "new-server")
	})

	t.Run("MarshalJSON round-trips empty config", func(t *testing.T) {
		config := &Config{}
		config.AddServer("test", Server{Command: "/bin/app", Args: []string{"mcp", "start"}})

		out, err := json.Marshal(config)
		require.NoError(t, err)

		config2 := &Config{}
		err = json.Unmarshal(out, config2)
		require.NoError(t, err)

		assert.True(t, config2.HasServer("test"))
		assert.Equal(t, "/bin/app", config2.ContextServers["test"].Command)
		assert.Equal(t, []string{"mcp", "start"}, config2.ContextServers["test"].Args)
	})

	t.Run("MarshalJSON omits context_servers when nil", func(t *testing.T) {
		data := []byte(`{"theme":"One Dark"}`)

		config := &Config{}
		err := json.Unmarshal(data, config)
		require.NoError(t, err)

		out, err := json.Marshal(config)
		require.NoError(t, err)

		var result map[string]json.RawMessage
		err = json.Unmarshal(out, &result)
		require.NoError(t, err)

		assert.NotContains(t, result, "context_servers")
		assert.Contains(t, result, "theme")
	})

	t.Run("RemoveServer then marshal does not include removed server", func(t *testing.T) {
		data := []byte(`{
			"theme": "Gruvbox",
			"context_servers": {
				"server-a": {"command": "/bin/a"},
				"server-b": {"command": "/bin/b"}
			}
		}`)

		config := &Config{}
		err := json.Unmarshal(data, config)
		require.NoError(t, err)

		config.RemoveServer("server-a")

		out, err := json.Marshal(config)
		require.NoError(t, err)

		config2 := &Config{}
		err = json.Unmarshal(out, config2)
		require.NoError(t, err)

		assert.False(t, config2.HasServer("server-a"))
		assert.True(t, config2.HasServer("server-b"))

		var result map[string]json.RawMessage
		err = json.Unmarshal(out, &result)
		require.NoError(t, err)
		assert.Contains(t, result, "theme")
	})
}

func TestServer(t *testing.T) {
	t.Run("Server with command only", func(t *testing.T) {
		server := Server{
			Command: "/usr/local/bin/myapp",
		}

		assert.Equal(t, "/usr/local/bin/myapp", server.Command)
		assert.Empty(t, server.Args)
		assert.Empty(t, server.Env)
		assert.Empty(t, server.URL)
	})

	t.Run("Server with args", func(t *testing.T) {
		server := Server{
			Command: "/usr/local/bin/myapp",
			Args:    []string{"mcp", "start", "--log-level", "debug"},
		}

		assert.Equal(t, 4, len(server.Args))
		assert.Equal(t, "mcp", server.Args[0])
		assert.Equal(t, "debug", server.Args[3])
	})

	t.Run("Server with environment variables", func(t *testing.T) {
		server := Server{
			Command: "/usr/local/bin/myapp",
			Env: map[string]string{
				"DEBUG":    "true",
				"LOG_FILE": "/var/log/app.log",
			},
		}

		assert.Equal(t, 2, len(server.Env))
		assert.Equal(t, "true", server.Env["DEBUG"])
		assert.Equal(t, "/var/log/app.log", server.Env["LOG_FILE"])
	})

	t.Run("Server with remote URL", func(t *testing.T) {
		server := Server{
			URL: "https://api.example.com/mcp",
			Headers: map[string]string{
				"Authorization": "Bearer token123",
			},
		}

		assert.Equal(t, "https://api.example.com/mcp", server.URL)
		assert.Equal(t, 1, len(server.Headers))
		assert.Equal(t, "Bearer token123", server.Headers["Authorization"])
		assert.Empty(t, server.Command)
	})

	t.Run("Server with all fields", func(t *testing.T) {
		server := Server{
			Command: "/usr/local/bin/myapp",
			Args:    []string{"mcp", "start"},
			Env: map[string]string{
				"API_KEY": "secret",
			},
		}

		assert.NotEmpty(t, server.Command)
		assert.NotEmpty(t, server.Args)
		assert.NotEmpty(t, server.Env)
	})
}

func TestPreprocess(t *testing.T) {
	t.Run("Preprocess strips line comments", func(t *testing.T) {
		data := []byte(`{
			// This is a comment
			"theme": "One Dark"
		}`)

		config := &Config{}
		result, err := config.Preprocess(data)
		require.NoError(t, err)

		var out map[string]any
		err = json.Unmarshal(result, &out)
		require.NoError(t, err)
		assert.Equal(t, "One Dark", out["theme"])
	})

	t.Run("Preprocess strips trailing commas in objects", func(t *testing.T) {
		data := []byte(`{
			"theme": "One Dark",
			"font_size": 14,
		}`)

		config := &Config{}
		result, err := config.Preprocess(data)
		require.NoError(t, err)

		var out map[string]any
		err = json.Unmarshal(result, &out)
		require.NoError(t, err)
		assert.Equal(t, float64(14), out["font_size"])
	})

	t.Run("Preprocess strips trailing commas in arrays", func(t *testing.T) {
		data := []byte(`{
			"favorite_models": ["gpt-4", "claude",],
		}`)

		config := &Config{}
		result, err := config.Preprocess(data)
		require.NoError(t, err)

		var out map[string]any
		err = json.Unmarshal(result, &out)
		require.NoError(t, err)
		models := out["favorite_models"].([]any)
		assert.Equal(t, 2, len(models))
	})

	t.Run("Preprocess handles real Zed settings.json format", func(t *testing.T) {
		data := []byte(`// Zed settings
//
// For information on how to configure Zed, see the Zed
// documentation: https://zed.dev/docs/configuring-zed
{
  "agent": {
    "tool_permissions": {
      "default": "allow"
    },
    "default_model": {
      "provider": "copilot_chat",
      "model": "claude-sonnet-4",
    },
    "favorite_models": [],
  },
  "ui_font_size": 16,
  "theme": {
    "mode": "system",
    "light": "One Light",
    "dark": "One Dark",
  },
  "context_servers": {
    "my-server": {
      "command": "/usr/local/bin/myapp",
      "args": ["mcp", "start"]
    }
  },
}`)

		// Mirror what loadConfig does: Preprocess first, then json.Unmarshal.
		// json.Unmarshal validates the entire input before dispatching to a
		// custom UnmarshalJSON, so JSONC must be normalised to strict JSON
		// before it is handed to the standard library.
		config := &Config{}
		normalized, err := config.Preprocess(data)
		require.NoError(t, err)

		err = json.Unmarshal(normalized, config)
		require.NoError(t, err)

		assert.True(t, config.HasServer("my-server"))
		assert.Equal(t, "/usr/local/bin/myapp", config.ContextServers["my-server"].Command)
		// Other settings are preserved in extra
		assert.Contains(t, config.extra, "theme")
		assert.Contains(t, config.extra, "ui_font_size")
		assert.Contains(t, config.extra, "agent")
	})

	t.Run("Preprocess leaves valid JSON unchanged", func(t *testing.T) {
		data := []byte(`{"theme":"One Dark","font_size":14}`)

		config := &Config{}
		result, err := config.Preprocess(data)
		require.NoError(t, err)

		var out map[string]any
		err = json.Unmarshal(result, &out)
		require.NoError(t, err)
		assert.Equal(t, "One Dark", out["theme"])
		assert.Equal(t, float64(14), out["font_size"])
	})

	t.Run("Preprocess returns error for invalid content", func(t *testing.T) {
		data := []byte(`{ not valid at all ???`)

		config := &Config{}
		_, err := config.Preprocess(data)
		assert.Error(t, err)
	})
}
