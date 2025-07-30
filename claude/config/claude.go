// Package config provides utilities for managing Claude Desktop MCP server configuration.
// It handles reading, writing, and modifying the Claude configuration file that defines MCP servers.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the structure of Claude's desktop configuration
type Config struct {
	MCPServers map[string]MCPServer `json:"mcpServers"`
}

// MCPServer represents an MCP server configuration entry
type MCPServer struct {
	Command string            `json:"command"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
}

// Manager handles Claude MCP configuration file operations
type Manager struct {
	configPath string
}

// NewClaudeConfigManager creates a new config manager with the default or specified path
func NewClaudeConfigManager(configPath string) *Manager {
	if configPath == "" {
		configPath = getDefaultClaudeConfigPath()
	}
	return &Manager{
		configPath: configPath,
	}
}

// LoadConfig loads the Claude configuration from file
func (cm *Manager) LoadConfig() (*Config, error) {
	// Check if config file exists
	if _, err := os.Stat(cm.configPath); os.IsNotExist(err) {
		// Return empty config if file doesn't exist
		return &Config{
			MCPServers: make(map[string]MCPServer),
		}, nil
	}

	data, err := os.ReadFile(cm.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read Claude configuration file at '%s': %w", cm.configPath, err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse Claude configuration file at '%s': invalid JSON format: %w", cm.configPath, err)
	}

	// Initialize MCPServers map if it's nil
	if config.MCPServers == nil {
		config.MCPServers = make(map[string]MCPServer)
	}

	return &config, nil
}

// SaveConfig saves the Claude configuration to file
func (cm *Manager) SaveConfig(config *Config) error {
	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(cm.configPath), 0o755); err != nil {
		return fmt.Errorf("failed to create Claude configuration directory at '%s': %w", filepath.Dir(cm.configPath), err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal Claude configuration to JSON: %w", err)
	}

	if err := os.WriteFile(cm.configPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write Claude configuration file at '%s': %w", cm.configPath, err)
	}

	return nil
}

// AddServer adds or updates an MCP server configuration
func (cm *Manager) AddServer(name string, server MCPServer) error {
	config, err := cm.LoadConfig()
	if err != nil {
		return err
	}

	config.MCPServers[name] = server
	return cm.SaveConfig(config)
}

// RemoveServer removes an MCP server configuration
func (cm *Manager) RemoveServer(name string) error {
	config, err := cm.LoadConfig()
	if err != nil {
		return err
	}

	delete(config.MCPServers, name)
	return cm.SaveConfig(config)
}

// HasServer checks if a server with the given name exists
func (cm *Manager) HasServer(name string) (bool, error) {
	config, err := cm.LoadConfig()
	if err != nil {
		return false, err
	}

	_, exists := config.MCPServers[name]
	return exists, nil
}

// GetConfigPath returns the path to the Claude configuration file being used
func (cm *Manager) GetConfigPath() string {
	return cm.configPath
}

// BackupConfig creates a backup of the current configuration file
func (cm *Manager) BackupConfig() error {
	if _, err := os.Stat(cm.configPath); os.IsNotExist(err) {
		// No config file to backup
		return nil
	}

	backupPath := cm.configPath + ".backup"
	data, err := os.ReadFile(cm.configPath)
	if err != nil {
		return fmt.Errorf("failed to read Claude configuration file for backup at '%s': %w", cm.configPath, err)
	}

	if err := os.WriteFile(backupPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write backup configuration file at '%s': %w", backupPath, err)
	}

	return nil
}
