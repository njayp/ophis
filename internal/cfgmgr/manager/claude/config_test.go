package claude

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
		assert.Equal(t, 1, len(config.Servers))
	})

	t.Run("AddServer initializes map if nil", func(t *testing.T) {
		config := &Config{}
		assert.Nil(t, config.Servers)

		server := Server{
			Command: "/usr/local/bin/myapp",
		}

		config.AddServer("test", server)
		assert.NotNil(t, config.Servers)
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

		assert.Equal(t, 1, len(config.Servers))
		assert.Equal(t, 2, len(config.Servers["test"].Args))
	})

	t.Run("HasServer returns false for non-existent server", func(t *testing.T) {
		config := &Config{
			Servers: map[string]Server{
				"server1": {Command: "/bin/app"},
			},
		}

		assert.False(t, config.HasServer("non-existent"))
		assert.True(t, config.HasServer("server1"))
	})

	t.Run("RemoveServer removes existing server", func(t *testing.T) {
		config := &Config{
			Servers: map[string]Server{
				"server1": {Command: "/bin/app1"},
				"server2": {Command: "/bin/app2"},
			},
		}

		config.RemoveServer("server1")

		assert.False(t, config.HasServer("server1"))
		assert.True(t, config.HasServer("server2"))
		assert.Equal(t, 1, len(config.Servers))
	})

	t.Run("RemoveServer handles non-existent server", func(t *testing.T) {
		config := &Config{
			Servers: map[string]Server{
				"server1": {Command: "/bin/app"},
			},
		}

		config.RemoveServer("non-existent")

		assert.Equal(t, 1, len(config.Servers))
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

		assert.Equal(t, 3, len(config.Servers))
		for _, s := range servers {
			assert.True(t, config.HasServer(s.name))
		}

		config.RemoveServer("helm")
		assert.Equal(t, 2, len(config.Servers))
		assert.False(t, config.HasServer("helm"))
	})
}

func TestMCPServer(t *testing.T) {
	t.Run("Server with command only", func(t *testing.T) {
		server := Server{
			Command: "/usr/local/bin/myapp",
		}

		assert.Equal(t, "/usr/local/bin/myapp", server.Command)
		assert.Empty(t, server.Args)
		assert.Empty(t, server.Env)
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
