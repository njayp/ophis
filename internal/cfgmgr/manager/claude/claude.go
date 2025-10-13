package claude

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

// Platform handles VSCode MCP configuration file operations
type Platform struct {
	configPath string
}

// NewClaudeCodeConfigManager creates a new config manager with the default or specified path
func NewClaudeCodeConfigManager() *Platform {
	return &Platform{
		configPath: getDefaultClaudeConfigPath(),
	}
}

// ConfigPath returns the path to the VSCode configuration file being used
func (cm *Platform) ConfigPath() string {
	return cm.configPath
}

// AddServer adds the server to the provided config
func (cm *Platform) AddServer(config *Config, name string, server MCPServer) {
	config.MCPServers[name] = server
}

// RemoveServer adds the server to the provided config
func (cm *Platform) RemoveServer(config *Config, name string) {
	delete(config.MCPServers, name)
}
