package zed

import (
	"encoding/json"
	"fmt"
	"slices"

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

// Serialize implements manager.Serializer by surgically updating the
// "context_servers" key in the original JSONC AST and packing it back,
// preserving all comments, trailing commas, and formatting in the rest
// of the file. If original is empty (new file), it falls back to standard
// indented JSON.
func (c *Config) Serialize(original []byte) ([]byte, error) {
	if len(original) == 0 {
		return json.MarshalIndent(c, "", "  ")
	}

	root, err := hujson.Parse(original)
	if err != nil {
		// Unparseable original — fall back to standard JSON.
		return json.MarshalIndent(c, "", "  ")
	}

	obj, ok := root.Value.(*hujson.Object)
	if !ok {
		return json.MarshalIndent(c, "", "  ")
	}

	// Locate the index of the existing "context_servers" member, if any.
	csIdx := -1
	for i, m := range obj.Members {
		if lit, ok := m.Name.Value.(hujson.Literal); ok && lit.String() == "context_servers" {
			csIdx = i
			break
		}
	}

	if len(c.ContextServers) == 0 {
		// No context servers — remove the key entirely if present.
		if csIdx >= 0 {
			obj.Members = slices.Delete(obj.Members, csIdx, csIdx+1)
		}
		return root.Pack(), nil
	}

	// Marshal the current context_servers map to indented JSON using the
	// same 2-space indent as Zed's settings file, then parse it as a hujson
	// Value so it can be inserted into the existing JSONC AST.
	// prefix="  " ensures the closing brace and nested keys align correctly
	// when the key sits at the top level of the root object.
	csJSON, err := json.MarshalIndent(c.ContextServers, "  ", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal context_servers: %w", err)
	}
	csVal, err := hujson.Parse(csJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to parse context_servers JSON: %w", err)
	}

	if csIdx >= 0 {
		// Update the value in place, keeping the key's surrounding whitespace
		// and any comments that immediately follow it.
		obj.Members[csIdx].Value = csVal
	} else {
		// Append a new member. Give the key a leading newline + indent so it
		// sits on its own line consistent with the rest of the file.
		obj.Members = append(obj.Members, hujson.ObjectMember{
			Name: hujson.Value{
				BeforeExtra: hujson.Extra("\n  "),
				Value:       hujson.String("context_servers"),
			},
			Value: csVal,
		})
	}

	return root.Pack(), nil
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
