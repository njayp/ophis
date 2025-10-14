package claude

import "fmt"

// Config represents the structure of Claude Desktop's configuration file.
type Config struct {
	MCPServers map[string]MCPServer `json:"mcpServers"`
}

// AddServer adds or updates a server in the configuration.
func (c *Config) AddServer(name string, server MCPServer) {
	if c.MCPServers == nil {
		c.MCPServers = make(map[string]MCPServer)
	}
	c.MCPServers[name] = server
}

// HasServer returns true if a server with the given name exists in the configuration.
func (c *Config) HasServer(name string) bool {
	_, ok := c.MCPServers[name]
	return ok
}

// RemoveServer removes a server from the configuration.
func (c *Config) RemoveServer(name string) {
	delete(c.MCPServers, name)
}

// Print displays all configured MCP servers.
func (c *Config) Print() {
	if len(c.MCPServers) == 0 {
		fmt.Println("No MCP servers are currently configured.")
		return
	}

	for name, server := range c.MCPServers {
		fmt.Printf("Server: %s\n", name)
		server.Print()
		fmt.Println()
	}
}
