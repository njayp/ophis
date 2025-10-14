package vscode

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	t.Run("AddServer adds new server", func(t *testing.T) {
		config := &Config{}
		server := MCPServer{
			Type:    "stdio",
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

		server := MCPServer{
			Type:    "stdio",
			Command: "/usr/local/bin/myapp",
		}

		config.AddServer("test", server)
		assert.NotNil(t, config.Servers)
		assert.True(t, config.HasServer("test"))
	})

	t.Run("AddServer updates existing server", func(t *testing.T) {
		config := &Config{}

		originalServer := MCPServer{
			Type:    "stdio",
			Command: "/usr/local/bin/myapp",
			Args:    []string{"start"},
		}
		config.AddServer("test", originalServer)

		updatedServer := MCPServer{
			Type:    "stdio",
			Command: "/usr/local/bin/myapp",
			Args:    []string{"start", "--verbose"},
		}
		config.AddServer("test", updatedServer)

		assert.Equal(t, 1, len(config.Servers))
		assert.Equal(t, 2, len(config.Servers["test"].Args))
	})

	t.Run("HasServer returns false for non-existent server", func(t *testing.T) {
		config := &Config{
			Servers: map[string]MCPServer{
				"server1": {Command: "/bin/app"},
			},
		}

		assert.False(t, config.HasServer("non-existent"))
		assert.True(t, config.HasServer("server1"))
	})

	t.Run("RemoveServer removes existing server", func(t *testing.T) {
		config := &Config{
			Servers: map[string]MCPServer{
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
			Servers: map[string]MCPServer{
				"server1": {Command: "/bin/app"},
			},
		}

		config.RemoveServer("non-existent")

		assert.Equal(t, 1, len(config.Servers))
		assert.True(t, config.HasServer("server1"))
	})

	t.Run("Config with inputs", func(t *testing.T) {
		config := &Config{
			Inputs: []Input{
				{
					Type:        "promptString",
					ID:          "api-key",
					Description: "Enter API key",
					Password:    true,
				},
			},
			Servers: map[string]MCPServer{
				"server1": {Command: "/bin/app"},
			},
		}

		assert.Equal(t, 1, len(config.Inputs))
		assert.Equal(t, "api-key", config.Inputs[0].ID)
		assert.True(t, config.Inputs[0].Password)
	})

	t.Run("Multiple servers can be managed", func(t *testing.T) {
		config := &Config{}

		servers := []struct {
			name   string
			server MCPServer
		}{
			{"kubectl", MCPServer{Type: "stdio", Command: "/usr/local/bin/kubectl"}},
			{"helm", MCPServer{Type: "stdio", Command: "/usr/local/bin/helm"}},
			{"argocd", MCPServer{Type: "stdio", Command: "/usr/local/bin/argocd"}},
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
	t.Run("Server with stdio type", func(t *testing.T) {
		server := MCPServer{
			Type:    "stdio",
			Command: "/usr/local/bin/myapp",
		}

		assert.Equal(t, "stdio", server.Type)
		assert.Equal(t, "/usr/local/bin/myapp", server.Command)
		assert.Empty(t, server.Args)
		assert.Empty(t, server.Env)
		assert.Empty(t, server.URL)
	})

	t.Run("Server with args", func(t *testing.T) {
		server := MCPServer{
			Type:    "stdio",
			Command: "/usr/local/bin/myapp",
			Args:    []string{"mcp", "start", "--log-level", "debug"},
		}

		assert.Equal(t, 4, len(server.Args))
		assert.Equal(t, "mcp", server.Args[0])
		assert.Equal(t, "debug", server.Args[3])
	})

	t.Run("Server with environment variables", func(t *testing.T) {
		server := MCPServer{
			Type:    "stdio",
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

	t.Run("Server with HTTP type", func(t *testing.T) {
		server := MCPServer{
			Type: "http",
			URL:  "https://api.example.com/mcp",
			Headers: map[string]string{
				"Authorization": "Bearer token",
				"Content-Type":  "application/json",
			},
		}

		assert.Equal(t, "http", server.Type)
		assert.Equal(t, "https://api.example.com/mcp", server.URL)
		assert.Equal(t, 2, len(server.Headers))
		assert.Equal(t, "Bearer token", server.Headers["Authorization"])
	})

	t.Run("Server with all fields", func(t *testing.T) {
		server := MCPServer{
			Type:    "stdio",
			Command: "/usr/local/bin/myapp",
			Args:    []string{"mcp", "start"},
			Env: map[string]string{
				"API_KEY": "secret",
			},
		}

		assert.NotEmpty(t, server.Type)
		assert.NotEmpty(t, server.Command)
		assert.NotEmpty(t, server.Args)
		assert.NotEmpty(t, server.Env)
	})
}

func TestInput(t *testing.T) {
	t.Run("Input with password flag", func(t *testing.T) {
		input := Input{
			Type:        "promptString",
			ID:          "password",
			Description: "Enter password",
			Password:    true,
		}

		assert.Equal(t, "promptString", input.Type)
		assert.Equal(t, "password", input.ID)
		assert.True(t, input.Password)
	})

	t.Run("Input without password flag", func(t *testing.T) {
		input := Input{
			Type:        "promptString",
			ID:          "username",
			Description: "Enter username",
		}

		assert.Equal(t, "promptString", input.Type)
		assert.Equal(t, "username", input.ID)
		assert.False(t, input.Password)
	})

	t.Run("Multiple inputs", func(t *testing.T) {
		inputs := []Input{
			{
				Type:        "promptString",
				ID:          "api-key",
				Description: "API Key",
				Password:    true,
			},
			{
				Type:        "promptString",
				ID:          "region",
				Description: "AWS Region",
			},
		}

		assert.Equal(t, 2, len(inputs))
		assert.True(t, inputs[0].Password)
		assert.False(t, inputs[1].Password)
	})
}
