package bridge

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
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

func TestCreateToolFromCmd_RequiredFlags(t *testing.T) {
	// Create a test command with required flags
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy application",
	}

	// Add flags
	cmd.Flags().String("env", "", "Deployment environment")
	cmd.Flags().String("version", "", "Application version")
	cmd.Flags().String("region", "us-east-1", "AWS region")
	cmd.Flags().Bool("dry-run", false, "Perform a dry run")

	// Mark some flags as required
	err := cmd.MarkFlagRequired("env")
	require.NoError(t, err)
	err = cmd.MarkFlagRequired("version")
	require.NoError(t, err)

	// Create tool from command
	tool := Selector{}.CreateToolFromCmd(cmd)

	// Verify schema structure
	inputSchema := tool.InputSchema.(*jsonschema.Schema)
	flagsSchema := inputSchema.Properties["flags"]

	// Check that the required flags are in the schema's required array
	require.NotNil(t, flagsSchema.Required)
	assert.Contains(t, flagsSchema.Required, "env", "env flag should be marked as required")
	assert.Contains(t, flagsSchema.Required, "version", "version flag should be marked as required")
	assert.NotContains(t, flagsSchema.Required, "region", "region flag should not be required")
	assert.NotContains(t, flagsSchema.Required, "dry-run", "dry-run flag should not be required")

	// Verify all flags are still present in properties
	assert.Contains(t, flagsSchema.Properties, "env")
	assert.Contains(t, flagsSchema.Properties, "version")
	assert.Contains(t, flagsSchema.Properties, "region")
	assert.Contains(t, flagsSchema.Properties, "dry-run")
}

// This test demonstrates the required flag functionality with a real-world example
func TestRequiredFlagsExample(t *testing.T) {
	// Create a command that represents a database connection tool
	dbCmd := &cobra.Command{
		Use:   "connect",
		Short: "Connect to database",
		Long:  "Connect to a database with specified credentials",
	}

	// Add various flags
	dbCmd.Flags().String("host", "localhost", "Database host")
	dbCmd.Flags().Int("port", 5432, "Database port")
	dbCmd.Flags().String("user", "", "Database username")
	dbCmd.Flags().String("password", "", "Database password")
	dbCmd.Flags().String("database", "", "Database name")
	dbCmd.Flags().Bool("ssl", true, "Use SSL connection")

	// Mark critical flags as required
	require.NoError(t, dbCmd.MarkFlagRequired("user"))
	require.NoError(t, dbCmd.MarkFlagRequired("password"))
	require.NoError(t, dbCmd.MarkFlagRequired("database"))

	// Create tool from command
	selector := Selector{}
	tool := selector.CreateToolFromCmd(dbCmd)

	// Convert to JSON to see the schema structure
	schema := tool.InputSchema.(*jsonschema.Schema)
	schemaJSON, err := json.MarshalIndent(schema, "", "  ")
	require.NoError(t, err)

	// Print for demonstration (normally would not print in tests)
	if testing.Verbose() {
		fmt.Printf("Generated MCP Tool Schema:\n%s\n", schemaJSON)
	}

	// Verify the flags schema contains required array
	flagsSchema := schema.Properties["flags"]
	require.NotNil(t, flagsSchema.Required)
	require.Len(t, flagsSchema.Required, 3, "Should have 3 required flags")

	// Verify required flags
	require.Contains(t, flagsSchema.Required, "user")
	require.Contains(t, flagsSchema.Required, "password")
	require.Contains(t, flagsSchema.Required, "database")

	// Verify optional flags are not in required array
	require.NotContains(t, flagsSchema.Required, "host")
	require.NotContains(t, flagsSchema.Required, "port")
	require.NotContains(t, flagsSchema.Required, "ssl")
}
