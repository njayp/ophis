package claude

import "fmt"

// Config represents the structure of Claude Desktop's configuration file.
type Config struct {
	Servers map[string]Server `json:"mcpServers"`
}

// AddServer adds or updates a server in the configuration.
func (c *Config) AddServer(name string, server Server) {
	if c.Servers == nil {
		c.Servers = make(map[string]Server)
	}
	c.Servers[name] = server
}

// HasServer returns true if a server with the given name exists in the configuration.
func (c *Config) HasServer(name string) bool {
	_, ok := c.Servers[name]
	return ok
}

// RemoveServer removes a server from the configuration.
func (c *Config) RemoveServer(name string) {
	delete(c.Servers, name)
}

// Print displays all configured MCP servers.
func (c *Config) Print() {
	if len(c.Servers) == 0 {
		fmt.Println("No MCP servers are currently configured.")
		return
	}

	for name, server := range c.Servers {
		fmt.Printf("Server: %s\n", name)
		server.Print()
		fmt.Println()
	}
}
