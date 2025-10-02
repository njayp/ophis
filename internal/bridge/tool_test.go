package bridge

import (
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateToolFromCmd(t *testing.T) {
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

	// Mark one flag as required
	err := cmd.MarkFlagRequired("count")
	require.NoError(t, err)

	// Create tool from command with a selector that accepts all flags
	tool := Selector{}.CreateToolFromCmd(cmd)

	// Verify tool properties
	assert.Equal(t, "test", tool.Name)
	assert.Contains(t, tool.Description, "This is a test command")
	assert.Contains(t, tool.Description, "test file.txt --output result.txt")
	assert.NotNil(t, tool.InputSchema)

	// Verify schema structure
	inputSchema := tool.InputSchema.(*jsonschema.Schema)
	assert.Equal(t, "object", inputSchema.Type)
	require.NotNil(t, inputSchema.Properties)
	assert.Contains(t, inputSchema.Properties, "flags")
	assert.Contains(t, inputSchema.Properties, "args")

	// Verify flags schema
	flagsSchema := inputSchema.Properties["flags"]
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

	// Verify required flags
	require.Len(t, flagsSchema.Required, 1, "Should have 1 required flag")
	assert.Contains(t, flagsSchema.Required, "count", "count flag should be marked as required")

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
	name := toolName(grandchild)
	assert.Equal(t, "root_child_grandchild", name)
}

func TestGenerateToolDescription(t *testing.T) {
	t.Run("Long and Example", func(t *testing.T) {
		cmd1 := &cobra.Command{
			Use:     "cmd1",
			Short:   "Short description",
			Long:    "Long description of cmd1",
			Example: "cmd1 --help",
		}
		desc1 := toolDescription(cmd1)
		assert.Contains(t, desc1, "Long description of cmd1")
		assert.Contains(t, desc1, "Examples:\ncmd1 --help")
	})

	t.Run("Short only", func(t *testing.T) {
		cmd2 := &cobra.Command{
			Use:   "cmd2",
			Short: "Short description of cmd2",
		}
		desc2 := toolDescription(cmd2)
		assert.Equal(t, "Short description of cmd2", desc2)
	})
}
