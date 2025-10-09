package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/njayp/ophis/internal/cfgmgr"
)

// Config represents the structure of Cursor's MCP configuration
type Config struct {
	Inputs     []Input              `json:"inputs,omitempty"`
	MCPServers map[string]MCPServer `json:"mcpServers"`
}

// Input represents an input variable configuration
type Input struct {
	Type        string `json:"type"`
	ID          string `json:"id"`
	Description string `json:"description"`
	Password    bool   `json:"password,omitempty"`
}

// MCPServer represents an MCP server configuration entry for Cursor
type MCPServer struct {
	Type    string            `json:"type,omitempty"`
	Command string            `json:"command"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	URL     string            `json:"url,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

// Manager handles Cursor MCP configuration file operations
type Manager struct {
	configPath string
	configType Type
}

// Type represents the type of Cursor configuration
type Type int

const (
	// WorkspaceConfig represents workspace-specific configuration (.cursor/mcp.json)
	WorkspaceConfig Type = iota
	// UserConfig represents user-global configuration (mcp.json)
	UserConfig
)

// String returns the string representation of the config type
func (t Type) String() string {
	switch t {
	case WorkspaceConfig:
		return "workspace"
	case UserConfig:
		return "user"
	default:
		return "unknown"
	}
}

// NewCursorConfigManager creates a new config manager with the default or specified path
func NewCursorConfigManager(configPath string, configType Type) *Manager {
	if configPath == "" {
		switch configType {
		case WorkspaceConfig:
			configPath = getDefaultWorkspaceConfigPath()
		case UserConfig:
			configPath = getDefaultUserConfigPath()
		}
	}
	return &Manager{
		configPath: configPath,
		configType: configType,
	}
}

// LoadConfig loads the Cursor configuration from file
func (cm *Manager) LoadConfig() (*Config, error) {
	config := &Config{
		MCPServers: make(map[string]MCPServer),
	}

	if err := cfgmgr.LoadJSONConfig(cm.configPath, config); err != nil {
		return nil, fmt.Errorf("failed to load Cursor configuration: %w", err)
	}

	// Initialize Servers map if it's nil (for existing configs)
	if config.MCPServers == nil {
		config.MCPServers = make(map[string]MCPServer)
	}

	return config, nil
}

// SaveConfig saves the Cursor configuration to file
func (cm *Manager) SaveConfig(config *Config) error {
	if err := cfgmgr.SaveJSONConfig(cm.configPath, config); err != nil {
		return fmt.Errorf("failed to save Cursor configuration: %w", err)
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

// GetConfigPath returns the path to the Cursor configuration file being used
func (cm *Manager) GetConfigPath() string {
	return cm.configPath
}

// BackupConfig creates a backup of the current configuration file
func (cm *Manager) BackupConfig() error {
	return cfgmgr.BackupConfigFile(cm.configPath)
}

// getDefaultWorkspaceConfigPath returns the default workspace config path (.cursor/mcp.json)
func getDefaultWorkspaceConfigPath() string {
	workingDir, err := os.Getwd()
	if err != nil {
		// Fallback to current directory
		return filepath.Join(".cursor", "mcp.json")
	}
	return filepath.Join(workingDir, ".cursor", "mcp.json")
}

// getDefaultUserConfigPath returns the default user config path (mcp.json)
func getDefaultUserConfigPath() string {
	return getDefaultCursorUserConfigPath()
}
