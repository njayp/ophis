package main

import (
	"encoding/json"
	"os"
	"slices"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestTools(t *testing.T) {
	cmd := rootCmd()
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
	var tools []mcp.Tool
	if err := json.Unmarshal(data, &tools); err != nil {
		t.Fatalf("Failed to unmarshal mcp-tools.json: %v", err)
	}
	if len(tools) != 6 {
		t.Fatalf("Expected 6 tools, got %d", len(tools))
	}

	// Check that the tools have the expected names
	names := []string{}
	for _, tool := range tools {
		names = append(names, tool.Name)
	}

	expectedNames := []string{
		"kubectl_get",
		"kubectl_describe",
		"kubectl_logs",
		"kubectl_top_pod",
		"kubectl_top_node",
		"kubectl_explain",
	}

	for _, expectedName := range expectedNames {
		if !slices.Contains(names, expectedName) {
			t.Fatalf("Expected tool name %q not found in generated tools: %v", expectedName, names)
		}
	}

	// Clean up
	if err := os.Remove("mcp-tools.json"); err != nil {
		t.Logf("Warning: failed to remove mcp-tools.json: %v", err)
	}
}
