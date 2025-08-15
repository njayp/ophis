package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the structure of VSCode's MCP configuration
type Config struct {
	Inputs  []Input              `json:"inputs,omitempty"`
	Servers map[string]MCPServer `json:"servers"`
}

// Input represents an input variable configuration
type Input struct {
	Type        string `json:"type"`
	ID          string `json:"id"`
	Description string `json:"description"`
	Password    bool   `json:"password,omitempty"`
}

// MCPServer represents an MCP server configuration entry for VSCode
type MCPServer struct {
	Type    string            `json:"type,omitempty"`
	Command string            `json:"command"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	URL     string            `json:"url,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

// Manager handles VSCode MCP configuration file operations
type Manager struct {
	configPath string
	configType Type
}

// Type represents the type of VSCode configuration
type Type int

const (
	// WorkspaceConfig represents workspace-specific configuration (.vscode/mcp.json)
	WorkspaceConfig Type = iota
	// UserConfig represents user-global configuration (mcp.json)
	UserConfig
)

// NewVSCodeConfigManager creates a new config manager with the default or specified path
func NewVSCodeConfigManager(configPath string, configType Type) *Manager {
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

// LoadConfig loads the VSCode configuration from file
func (cm *Manager) LoadConfig() (*Config, error) {
	// Check if config file exists
	if _, err := os.Stat(cm.configPath); os.IsNotExist(err) {
		// Return empty config if file doesn't exist
		return &Config{
			Servers: make(map[string]MCPServer),
		}, nil
	}

	data, err := os.ReadFile(cm.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read VSCode configuration file at '%s': %w", cm.configPath, err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse VSCode MCP configuration file at '%s': invalid JSON format: %w", cm.configPath, err)
	}

	// Initialize Servers map if it's nil
	if config.Servers == nil {
		config.Servers = make(map[string]MCPServer)
	}

	return &config, nil
}

// SaveConfig saves the VSCode configuration to file
func (cm *Manager) SaveConfig(config *Config) error {
	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(cm.configPath), 0o755); err != nil {
		return fmt.Errorf("failed to create VSCode configuration directory at '%s': %w", filepath.Dir(cm.configPath), err)
	}

	dataToWrite, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal VSCode MCP configuration to JSON: %w", err)
	}

	if err := os.WriteFile(cm.configPath, dataToWrite, 0o644); err != nil {
		return fmt.Errorf("failed to write VSCode configuration file at '%s': %w", cm.configPath, err)
	}

	return nil
}

// AddServer adds or updates an MCP server configuration
func (cm *Manager) AddServer(name string, server MCPServer) error {
	config, err := cm.LoadConfig()
	if err != nil {
		return err
	}

	config.Servers[name] = server
	return cm.SaveConfig(config)
}

// RemoveServer removes an MCP server configuration
func (cm *Manager) RemoveServer(name string) error {
	config, err := cm.LoadConfig()
	if err != nil {
		return err
	}

	delete(config.Servers, name)
	return cm.SaveConfig(config)
}

// HasServer checks if a server with the given name exists
func (cm *Manager) HasServer(name string) (bool, error) {
	config, err := cm.LoadConfig()
	if err != nil {
		return false, err
	}

	_, exists := config.Servers[name]
	return exists, nil
}

// GetConfigPath returns the path to the VSCode configuration file being used
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
		return fmt.Errorf("failed to read VSCode configuration file for backup at '%s': %w", cm.configPath, err)
	}

	if err := os.WriteFile(backupPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write backup configuration file at '%s': %w", backupPath, err)
	}

	return nil
}

// getDefaultWorkspaceConfigPath returns the default workspace config path (.vscode/mcp.json)
func getDefaultWorkspaceConfigPath() string {
	workingDir, err := os.Getwd()
	if err != nil {
		// Fallback to current directory
		return filepath.Join(".vscode", "mcp.json")
	}
	return filepath.Join(workingDir, ".vscode", "mcp.json")
}

// getDefaultUserConfigPath returns the default user config path (mcp.json)
func getDefaultUserConfigPath() string {
	return getDefaultVSCodeUserConfigPath()
}
