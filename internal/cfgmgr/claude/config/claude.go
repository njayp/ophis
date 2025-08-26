package config

import (
	"fmt"

	"github.com/njayp/ophis/internal/cfgmgr"
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
	config := &Config{
		MCPServers: make(map[string]MCPServer),
	}

	if err := cfgmgr.LoadJSONConfig(cm.configPath, config); err != nil {
		return nil, fmt.Errorf("failed to load Claude configuration: %w", err)
	}

	// Initialize MCPServers map if it's nil (for existing configs)
	if config.MCPServers == nil {
		config.MCPServers = make(map[string]MCPServer)
	}

	return config, nil
}

// SaveConfig saves the Claude configuration to file
func (cm *Manager) SaveConfig(config *Config) error {
	if err := cfgmgr.SaveJSONConfig(cm.configPath, config); err != nil {
		return fmt.Errorf("failed to save Claude configuration: %w", err)
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
	return cfgmgr.BackupConfigFile(cm.configPath)
}
