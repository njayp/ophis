package examples

import (
	"encoding/json"
	"os"
	"slices"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

// TestTools runs `mcp tools` and checks the output tool names against expectedNames
func TestTools(t *testing.T, cmd *cobra.Command, expectedNames []string) {
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

		// Check that the tools have the expected names
		names := []string{}
		for _, tool := range tools {
			names = append(names, tool.Name)
		}

		for _, expectedName := range expectedNames {
			if !slices.Contains(names, expectedName) {
				t.Fatalf("Expected tool name %q not found in generated tools: %v", expectedName, names)
			}
		}
	})

	// Clean up
	if err := os.Remove("mcp-tools.json"); err != nil {
		t.Logf("Warning: failed to remove mcp-tools.json: %v", err)
	}
}
