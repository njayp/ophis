package claude

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	t.Run("AddServer adds new server", func(t *testing.T) {
		config := &Config{}
		server := MCPServer{
			Command: "/usr/local/bin/myapp",
			Args:    []string{"mcp", "start"},
		}

		config.AddServer("test", server)

		assert.True(t, config.HasServer("test"))
		assert.Equal(t, 1, len(config.MCPServers))
	})

	t.Run("AddServer initializes map if nil", func(t *testing.T) {
		config := &Config{}
		assert.Nil(t, config.MCPServers)

		server := MCPServer{
			Command: "/usr/local/bin/myapp",
		}

		config.AddServer("test", server)
		assert.NotNil(t, config.MCPServers)
		assert.True(t, config.HasServer("test"))
	})

	t.Run("AddServer updates existing server", func(t *testing.T) {
		config := &Config{}

		originalServer := MCPServer{
			Command: "/usr/local/bin/myapp",
			Args:    []string{"start"},
		}
		config.AddServer("test", originalServer)

		updatedServer := MCPServer{
			Command: "/usr/local/bin/myapp",
			Args:    []string{"start", "--verbose"},
		}
		config.AddServer("test", updatedServer)

		assert.Equal(t, 1, len(config.MCPServers))
		assert.Equal(t, 2, len(config.MCPServers["test"].Args))
	})

	t.Run("HasServer returns false for non-existent server", func(t *testing.T) {
		config := &Config{
			MCPServers: map[string]MCPServer{
				"server1": {Command: "/bin/app"},
			},
		}

		assert.False(t, config.HasServer("non-existent"))
		assert.True(t, config.HasServer("server1"))
	})

	t.Run("RemoveServer removes existing server", func(t *testing.T) {
		config := &Config{
			MCPServers: map[string]MCPServer{
				"server1": {Command: "/bin/app1"},
				"server2": {Command: "/bin/app2"},
			},
		}

		config.RemoveServer("server1")

		assert.False(t, config.HasServer("server1"))
		assert.True(t, config.HasServer("server2"))
		assert.Equal(t, 1, len(config.MCPServers))
	})

	t.Run("RemoveServer handles non-existent server", func(t *testing.T) {
		config := &Config{
			MCPServers: map[string]MCPServer{
				"server1": {Command: "/bin/app"},
			},
		}

		config.RemoveServer("non-existent")

		assert.Equal(t, 1, len(config.MCPServers))
		assert.True(t, config.HasServer("server1"))
	})

	t.Run("Multiple servers can be managed", func(t *testing.T) {
		config := &Config{}

		servers := []struct {
			name   string
			server MCPServer
		}{
			{"kubectl", MCPServer{Command: "/usr/local/bin/kubectl"}},
			{"helm", MCPServer{Command: "/usr/local/bin/helm"}},
			{"argocd", MCPServer{Command: "/usr/local/bin/argocd"}},
		}

		for _, s := range servers {
			config.AddServer(s.name, s.server)
		}

		assert.Equal(t, 3, len(config.MCPServers))
		for _, s := range servers {
			assert.True(t, config.HasServer(s.name))
		}

		config.RemoveServer("helm")
		assert.Equal(t, 2, len(config.MCPServers))
		assert.False(t, config.HasServer("helm"))
	})
}

func TestMCPServer(t *testing.T) {
	t.Run("Server with command only", func(t *testing.T) {
		server := MCPServer{
			Command: "/usr/local/bin/myapp",
		}

		assert.Equal(t, "/usr/local/bin/myapp", server.Command)
		assert.Empty(t, server.Args)
		assert.Empty(t, server.Env)
	})

	t.Run("Server with args", func(t *testing.T) {
		server := MCPServer{
			Command: "/usr/local/bin/myapp",
			Args:    []string{"mcp", "start", "--log-level", "debug"},
		}

		assert.Equal(t, 4, len(server.Args))
		assert.Equal(t, "mcp", server.Args[0])
		assert.Equal(t, "debug", server.Args[3])
	})

	t.Run("Server with environment variables", func(t *testing.T) {
		server := MCPServer{
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

	t.Run("Server with all fields", func(t *testing.T) {
		server := MCPServer{
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
