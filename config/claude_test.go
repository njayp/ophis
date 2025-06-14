package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestClaudeConfigManager(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "claude_config_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "claude_desktop_config.json")
	manager := NewClaudeConfigManager(configPath)

	// Test loading empty config
	config, err := manager.LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load empty config: %v", err)
	}
	if config.MCPServers == nil {
		t.Fatal("MCPServers should be initialized")
	}
	if len(config.MCPServers) != 0 {
		t.Fatal("Empty config should have no servers")
	}

	// Test adding a server
	testServer := MCPServer{
		Command: "/path/to/test/executable",
		Args:    []string{"mcp", "start", "--log-level", "info"},
		Env:     map[string]string{"TEST_VAR": "test_value"},
	}

	err = manager.AddServer("test-server", testServer)
	if err != nil {
		t.Fatalf("Failed to add server: %v", err)
	}

	// Test loading config with server
	config, err = manager.LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config after adding server: %v", err)
	}
	if len(config.MCPServers) != 1 {
		t.Fatal("Config should have one server")
	}

	server, exists := config.MCPServers["test-server"]
	if !exists {
		t.Fatal("test-server should exist")
	}
	if server.Command != testServer.Command {
		t.Fatalf("Expected command %s, got %s", testServer.Command, server.Command)
	}
	if len(server.Args) != len(testServer.Args) {
		t.Fatalf("Expected %d args, got %d", len(testServer.Args), len(server.Args))
	}
	for i, arg := range testServer.Args {
		if server.Args[i] != arg {
			t.Fatalf("Expected arg[%d] %s, got %s", i, arg, server.Args[i])
		}
	}

	// Test HasServer
	hasServer, err := manager.HasServer("test-server")
	if err != nil {
		t.Fatalf("Failed to check if server exists: %v", err)
	}
	if !hasServer {
		t.Fatal("test-server should exist")
	}

	hasServer, err = manager.HasServer("non-existent")
	if err != nil {
		t.Fatalf("Failed to check if non-existent server exists: %v", err)
	}
	if hasServer {
		t.Fatal("non-existent server should not exist")
	}

	// Test removing server
	err = manager.RemoveServer("test-server")
	if err != nil {
		t.Fatalf("Failed to remove server: %v", err)
	}

	// Verify server is removed
	config, err = manager.LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config after removing server: %v", err)
	}
	if len(config.MCPServers) != 0 {
		t.Fatal("Config should have no servers after removal")
	}
}

func TestDefaultConfigPaths(t *testing.T) {
	// Test that getDefaultClaudeConfigPath returns a non-empty string
	path := getDefaultClaudeConfigPath()
	if path == "" {
		t.Fatal("Default config path should not be empty")
	}
	if !filepath.IsAbs(path) {
		t.Fatal("Default config path should be absolute")
	}
}

func TestClaudeConfigJSON(t *testing.T) {
	// Test JSON marshaling/unmarshaling
	config := &ClaudeConfig{
		MCPServers: map[string]MCPServer{
			"test": {
				Command: "/usr/bin/test",
				Args:    []string{"arg1", "arg2"},
				Env:     map[string]string{"KEY": "value"},
			},
		},
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	var unmarshaled ClaudeConfig
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal config: %v", err)
	}

	if len(unmarshaled.MCPServers) != 1 {
		t.Fatal("Unmarshaled config should have one server")
	}

	server := unmarshaled.MCPServers["test"]
	if server.Command != config.MCPServers["test"].Command {
		t.Fatal("Command not preserved in JSON round-trip")
	}
}
