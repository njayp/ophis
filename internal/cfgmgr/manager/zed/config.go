package zed

import (
	"encoding/json"
	"fmt"

	"github.com/tailscale/hujson"
)

// Config represents Zed's settings.json file.
// It uses custom JSON marshaling to preserve all existing Zed settings
// (e.g. theme, font, keybindings) while only managing the "context_servers" key.
type Config struct {
	ContextServers map[string]Server
	extra          map[string]json.RawMessage
}

// UnmarshalJSON deserializes the Zed settings file, extracting "context_servers"
// and preserving all other keys unchanged.
func (c *Config) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if cs, ok := raw["context_servers"]; ok {
		if err := json.Unmarshal(cs, &c.ContextServers); err != nil {
			return err
		}
		delete(raw, "context_servers")
	}

	c.extra = raw
	return nil
}

// Preprocess implements manager.Preprocessor by normalizing Zed's JSONC
// settings file to standard JSON before unmarshaling. Zed's settings.json
// uses JSONC syntax — it may contain // line comments, /* block comments */,
// and trailing commas after the last element in objects and arrays, none of
// which are valid in strict JSON.
func (c *Config) Preprocess(data []byte) ([]byte, error) {
	standardized, err := hujson.Standardize(data)
	if err != nil {
		return nil, fmt.Errorf("failed to normalize JSONC: %w", err)
	}
	return standardized, nil
}

// MarshalJSON serializes the config back to JSON, merging the managed
// "context_servers" with all other preserved Zed settings.
func (c Config) MarshalJSON() ([]byte, error) {
	result := make(map[string]json.RawMessage, len(c.extra)+1)

	for k, v := range c.extra {
		result[k] = v
	}

	if c.ContextServers != nil {
		cs, err := json.Marshal(c.ContextServers)
		if err != nil {
			return nil, err
		}
		result["context_servers"] = cs
	}

	return json.Marshal(result)
}

// AddServer adds or updates a server in the configuration.
func (c *Config) AddServer(name string, server Server) {
	if c.ContextServers == nil {
		c.ContextServers = make(map[string]Server)
	}
	c.ContextServers[name] = server
}

// HasServer returns true if a server with the given name exists in the configuration.
func (c *Config) HasServer(name string) bool {
	_, ok := c.ContextServers[name]
	return ok
}

// RemoveServer removes a server from the configuration.
func (c *Config) RemoveServer(name string) {
	delete(c.ContextServers, name)
}

// Print displays all configured MCP servers.
func (c *Config) Print() {
	if len(c.ContextServers) == 0 {
		fmt.Println("No MCP servers are currently configured.")
		return
	}

	for name, server := range c.ContextServers {
		fmt.Printf("Server: %s\n", name)
		server.Print()
		fmt.Println()
	}
}
