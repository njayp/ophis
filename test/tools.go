package test

import (
	"encoding/json"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

// GetTools runs `mcp tools` command and returns the parsed list of tools
// It fails if ophis.Command is not a root level subcommand
func GetTools(t *testing.T, cmd *cobra.Command) []*mcp.Tool {
	cmd.SetArgs([]string{"mcp", "tools"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Failed to generate tools: %v", err)
	}

	// Verify the mcp-tools.json file was created and contains valid JSON
	data, err := os.ReadFile("mcp-tools.json")
	if err != nil {
		t.Fatalf("Failed to read mcp-tools.json: %v", err)
	}

	var tools []*mcp.Tool
	if err := json.Unmarshal(data, &tools); err != nil {
		t.Fatalf("Failed to unmarshal mcp-tools.json: %v", err)
	}

	// Clean up the generated file
	if err := os.Remove("mcp-tools.json"); err != nil {
		t.Logf("Warning: failed to remove mcp-tools.json: %v", err)
	}

	return tools
}

// GetInputSchema extracts and returns the input schema from a tool
func GetInputSchema(t *testing.T, tool *mcp.Tool) *jsonschema.Schema {
	if tool.InputSchema == nil {
		t.Fatalf("Tool %q has no input schema", tool.Name)
	}

	data, err := json.Marshal(tool.InputSchema)
	if err != nil {
		t.Fatalf("Tool %q: failed to marshal input schema: %v", tool.Name, err)
	}

	schema := &jsonschema.Schema{}
	if err := json.Unmarshal(data, schema); err != nil {
		t.Fatalf("Tool %q: failed to unmarshal input schema into *jsonschema.Schema: %v (type was %T)",
			tool.Name, err, tool.InputSchema)
	}

	return schema
}

// ToolNames checks that the provided tools match the expectedNames
func ToolNames(t *testing.T, tools []*mcp.Tool, expectedNames ...string) {
	if len(tools) != len(expectedNames) {
		t.Errorf("expected %v tools, got %v", len(expectedNames), len(tools))
	}

	for _, tool := range tools {
		if !slices.Contains(expectedNames, tool.Name) {
			t.Errorf("Tool name not expected: %q", tool.Name)
		}
	}
}

// CmdPathsToToolNames converts command names with spaces to tool names with underscores
func CmdPathsToToolNames(paths []string) []string {
	names := make([]string, 0, len(paths))
	for _, path := range paths {
		names = append(names, strings.ReplaceAll(path, " ", "_"))
	}

	return names
}

// Tools runs `mcp tools` and checks the output tool names against expectedNames
// It fails if ophis.Command is not a root level subcommand
func Tools(t *testing.T, cmd *cobra.Command, expectedNames ...string) {
	tools := GetTools(t, cmd)

	t.Run("Expected Tools", func(t *testing.T) {
		ToolNames(t, tools, expectedNames...)
	})

	t.Run("Valid JSON Schema", func(t *testing.T) {
		// Check that each tool's input schema is valid JSON Schema
		for _, tool := range tools {
			if tool.InputSchema != nil {
				schema := GetInputSchema(t, tool)
				if schema.Type != "object" {
					t.Errorf("Tool %q: expected schema type 'object', got %q", tool.Name, schema.Type)
				}

				// Validate "flags" property if present
				if prop, ok := schema.Properties["flags"]; ok {
					if prop.Type != "object" {
						t.Errorf("Tool %q: expected 'flags' property type 'object', got %q", tool.Name, prop.Type)
					}
				}

				// Validate "args" property if present
				if prop, ok := schema.Properties["args"]; ok {
					if prop.Type != "array" || prop.Items.Type != "string" {
						t.Errorf("Tool %q: expected 'args' property type 'array', got %q, %q", tool.Name, prop.Type, prop.Items.Type)
					}
				}
			}
		}
	})
}
