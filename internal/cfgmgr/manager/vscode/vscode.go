package vscode

import (
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

// Platform handles VSCode MCP configuration file operations
type Platform struct {
	configPath string
}

// NewVSCodeConfigManager creates a new config manager with the default or specified path
func NewVSCodeConfigManager(workspace bool) *Platform {
	if workspace {
		return &Platform{
			configPath: getDefaultWorkspaceConfigPath(),
		}
	}

	return &Platform{
		configPath: getDefaultUserConfigPath(),
	}
}

// ConfigPath returns the path to the VSCode configuration file being used
func (cm *Platform) ConfigPath() string {
	return cm.configPath
}

// AddServer adds the server to the provided config
func (cm *Platform) AddServer(config *Config, name string, server MCPServer) {
	config.Servers[name] = server
}

// RemoveServer adds the server to the provided config
func (cm *Platform) RemoveServer(config *Config, name string) {
	delete(config.Servers, name)
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
