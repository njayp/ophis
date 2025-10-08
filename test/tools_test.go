package test

import (
	"testing"

	"github.com/njayp/ophis"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// createTestCommand creates a simple test command tree for testing.
func createTestCommand() *cobra.Command {
	root := &cobra.Command{
		Use:   "testcli",
		Short: "Test CLI",
	}

	get := &cobra.Command{
		Use:   "get [resource]",
		Short: "Get a resource",
		Run:   func(_ *cobra.Command, _ []string) {},
	}
	get.Flags().String("output", "json", "Output format")
	get.Flags().Bool("verbose", false, "Verbose output")

	list := &cobra.Command{
		Use:   "list",
		Short: "List resources",
		Run:   func(_ *cobra.Command, _ []string) {},
	}
	list.Flags().Int("limit", 10, "Limit results")

	root.AddCommand(get, list)
	root.AddCommand(ophis.Command(nil))

	return root
}

func TestGetTools(t *testing.T) {
	cmd := createTestCommand()

	// Test successful tool generation
	tools := GetTools(t, cmd)

	// Should have two tools: get and list
	assert.Len(t, tools, 2, "Expected 2 tools")

	// Verify ToolNames helper
	ToolNames(t, tools, "testcli_get", "testcli_list")

	// Verify each tool has required properties
	for _, tool := range tools {
		assert.NotEmpty(t, tool.Name, "Tool should have a name")
		assert.NotEmpty(t, tool.Description, "Tool should have a description")
		assert.NotNil(t, tool.InputSchema, "Tool should have an input schema")

		// Test GetInputSchema
		schema := GetInputSchema(t, tool)
		assert.NotNil(t, schema, "Should return a schema")
		assert.Equal(t, "object", schema.Type, "Schema type should be object")
	}
}

func TestCmdNamesToToolNames(t *testing.T) {
	cmdNames := []string{"get", "list all", "create this item"}
	expectedToolNames := []string{"get", "list_all", "create_this_item"}

	toolNames := CmdPathsToToolNames(cmdNames)
	assert.Equal(t, expectedToolNames, toolNames, "Tool names should match expected format")
}
