package bridge

import (
	"testing"

	"github.com/njayp/ophis/internal/test"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateToolFromCmd_Basic(t *testing.T) {
	// Create a simple test command
	cmd := &cobra.Command{
		Use:     "test [file]",
		Short:   "Test command",
		Long:    "This is a test command for testing the bridge package",
		Example: "test file.txt --output result.txt",
	}

	// Add some flags
	cmd.Flags().String("output", "", "Output file")
	cmd.Flags().Bool("verbose", false, "Verbose output")
	cmd.Flags().StringSlice("include", []string{}, "Include patterns")
	cmd.Flags().Int("count", 10, "Number of items")

	// Create tool from command
	tool := CreateToolFromCmd(cmd, nil)

	// Verify tool properties
	assert.Equal(t, "test", tool.Name)
	assert.Contains(t, tool.Description, "This is a test command")
	assert.Contains(t, tool.Description, "test file.txt --output result.txt")
	assert.NotNil(t, tool.InputSchema)

	// Verify the tool passes MCP schema validation
	err := test.ValidateToolSchema(tool)
	assert.NoError(t, err, "Tool should pass MCP schema validation")

	// Verify schema structure
	require.NotNil(t, tool.InputSchema.Properties)
	assert.Contains(t, tool.InputSchema.Properties, "flags")
	assert.Contains(t, tool.InputSchema.Properties, "args")

	// Verify flags schema
	flagsSchema := tool.InputSchema.Properties["flags"]
	require.NotNil(t, flagsSchema.Properties)
	assert.Contains(t, flagsSchema.Properties, "output")
	assert.Contains(t, flagsSchema.Properties, "verbose")
	assert.Contains(t, flagsSchema.Properties, "include")
	assert.Contains(t, flagsSchema.Properties, "count")

	// Verify flag types
	assert.Equal(t, "string", flagsSchema.Properties["output"].Type)
	assert.Equal(t, "boolean", flagsSchema.Properties["verbose"].Type)
	assert.Equal(t, "array", flagsSchema.Properties["include"].Type)
	assert.Equal(t, "integer", flagsSchema.Properties["count"].Type)

	// Verify array items schema
	includeSchema := flagsSchema.Properties["include"]
	assert.NotNil(t, includeSchema.Items)
	assert.Equal(t, "string", includeSchema.Items.Type)
}

func TestGenerateToolName(t *testing.T) {
	root := &cobra.Command{
		Use: "root",
	}
	child := &cobra.Command{
		Use: "child",
	}
	grandchild := &cobra.Command{
		Use: "grandchild",
	}

	root.AddCommand(child)
	child.AddCommand(grandchild)
	name := generateToolName(grandchild)
	assert.Equal(t, "root_child_grandchild", name)
}
