package main

import (
	"encoding/json"
	"os"
	"slices"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestTools(t *testing.T) {
	cmd := createMakeCommands()
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

	t.Run("Expected Tools", func(t *testing.T) {
		var tools []mcp.Tool
		if err := json.Unmarshal(data, &tools); err != nil {
			t.Fatalf("Failed to unmarshal mcp-tools.json: %v", err)
		}
		if len(tools) != 2 {
			t.Fatalf("Expected 2 tools, got %d", len(tools))
		}

		// Check that the tools have the expected names
		names := []string{tools[0].Name, tools[1].Name}
		if !slices.Contains(names, "make_test") || !slices.Contains(names, "make_lint") {
			t.Fatalf("Unexpected tool names: %v", names)
		}
	})

	// Clean up
	if err := os.Remove("mcp-tools.json"); err != nil {
		t.Logf("Warning: failed to remove mcp-tools.json: %v", err)
	}
}
