package ophis

import (
	"slices"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// buildCommandTree creates a command tree from a list of command names.
// The first command becomes the root, and subsequent commands are nested.
func buildCommandTree(names ...string) *cobra.Command {
	if len(names) == 0 {
		return nil
	}

	root := &cobra.Command{Use: names[0]}
	parent := root

	for _, name := range names[1:] {
		child := &cobra.Command{
			Use: name,
			Run: func(_ *cobra.Command, _ []string) {},
		}

		parent.AddCommand(child)
		parent = child
	}

	return parent
}

type SomeJSONObject struct {
	Foo    string
	Bar    int
	FooBar struct {
		Baz string
	}
}

type SomeJSONArray []SomeJSONObject

func TestCreateToolFromCmd(t *testing.T) {
	// Create a simple test command
	cmd := &cobra.Command{
		Use:     "test [file]",
		Short:   "Test command",
		Long:    "This is a test command for testing the ophis package",
		Example: "test file.txt --output result.txt",
	}

	// Add some flags
	cmd.Flags().String("output", "", "Output file")
	cmd.Flags().Bool("verbose", false, "Verbose output")
	cmd.Flags().IntSlice("include", []int{}, "Include patterns")
	cmd.Flags().StringSlice("greeting", []string{"hello", "world"}, "Include patterns")
	cmd.Flags().Int("count", 10, "Number of items")
	cmd.Flags().StringToString("labels", map[string]string{"hello": "world", "go": "lang"}, "Key-value labels")
	cmd.Flags().StringToInt("ports", map[string]int{"life": 42, "power": 9001}, "Port mappings")

	// generate schema for a test object
	aJSONObjSchema, err := jsonschema.For[SomeJSONObject](nil)
	require.NoError(t, err)
	bytes, err := aJSONObjSchema.MarshalJSON()
	require.NoError(t, err)

	// now create flag that has a json schema that represents a json object
	cmd.Flags().String("a_json_obj", "", "Some JSON Object")
	jsonobj := cmd.Flags().Lookup("a_json_obj")
	jsonobj.Annotations = make(map[string][]string)
	jsonobj.Annotations["jsonschema"] = []string{string(bytes)}

	// generate schema for a test array
	aJSONArraySchema, err := jsonschema.For[SomeJSONArray](nil)
	require.NoError(t, err)
	bytes, err = aJSONArraySchema.MarshalJSON()
	require.NoError(t, err)

	// now create flag that has a json schema that represents a json array
	// note that we can supply a default for the flag here but it's not mapped to the schema default
	cmd.Flags().String("a_json_array", "[]", "Some JSON Array")
	jsonarray := cmd.Flags().Lookup("a_json_array")
	jsonarray.Annotations = make(map[string][]string)
	jsonarray.Annotations["jsonschema"] = []string{string(bytes)}

	// Add a hidden flag
	cmd.Flags().String("hidden", "secret", "Hidden flag")
	err = cmd.Flags().MarkHidden("hidden")
	require.NoError(t, err)

	// Add a deprecated flag
	cmd.Flags().String("old", "", "Old flag")
	err = cmd.Flags().MarkDeprecated("old", "Use --new instead")
	require.NoError(t, err)

	// Mark one flag as required
	err = cmd.MarkFlagRequired("count")
	require.NoError(t, err)

	parent := &cobra.Command{
		Use:   "parent",
		Short: "Parent command",
	}

	// add persistent flag to parent
	parent.PersistentFlags().String("config", "", "Config file")
	parent.AddCommand(cmd)

	t.Run("Default Selector", func(t *testing.T) {
		// Create tool from command with a selector that accepts all flags
		tool := Selector{}.createToolFromCmd(cmd, "parent")

		// Verify tool properties
		assert.Equal(t, "parent_test", tool.Name)
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
		assert.Contains(t, flagsSchema.Properties, "greeting")
		assert.Contains(t, flagsSchema.Properties, "labels")
		assert.Contains(t, flagsSchema.Properties, "ports")
		assert.Contains(t, flagsSchema.Properties, "a_json_obj")
		assert.Contains(t, flagsSchema.Properties, "a_json_array")

		// Verify excluded flags
		assert.NotContains(t, flagsSchema.Properties, "hidden", "Should not include hidden flag")
		assert.NotContains(t, flagsSchema.Properties, "old", "Should not include deprecated flag")

		// Verify flag types
		assert.Equal(t, "string", flagsSchema.Properties["output"].Type)
		assert.Equal(t, "boolean", flagsSchema.Properties["verbose"].Type)
		assert.Equal(t, "array", flagsSchema.Properties["include"].Type)
		assert.Equal(t, "integer", flagsSchema.Properties["count"].Type)
		assert.Equal(t, "array", flagsSchema.Properties["greeting"].Type)
		assert.Equal(t, "object", flagsSchema.Properties["labels"].Type)
		assert.Equal(t, "object", flagsSchema.Properties["ports"].Type)
		assert.Equal(t, "object", flagsSchema.Properties["a_json_obj"].Type)
		assert.Equal(t, "array", flagsSchema.Properties["a_json_array"].Type)

		// Verify required flags
		require.Len(t, flagsSchema.Required, 1, "Should have 1 required flag")
		assert.Contains(t, flagsSchema.Required, "count", "count flag should be marked as required")

		// Verify default values
		assert.NotNil(t, flagsSchema.Properties["verbose"].Default)
		assert.JSONEq(t, "false", string(flagsSchema.Properties["verbose"].Default))
		assert.NotNil(t, flagsSchema.Properties["count"].Default)
		assert.JSONEq(t, "10", string(flagsSchema.Properties["count"].Default))
		assert.NotNil(t, flagsSchema.Properties["greeting"].Default)
		assert.JSONEq(t, `["hello","world"]`, string(flagsSchema.Properties["greeting"].Default))
		assert.JSONEq(t, `{"life":42, "power":9001}`, string(flagsSchema.Properties["ports"].Default))
		assert.JSONEq(t, `{"hello":"world", "go":"lang"}`, string(flagsSchema.Properties["labels"].Default))
		// Empty string and empty array should not have defaults set
		assert.Nil(t, flagsSchema.Properties["output"].Default)
		assert.Nil(t, flagsSchema.Properties["include"].Default)

		// json schema defaults are not populated
		assert.Nil(t, flagsSchema.Properties["a_json_obj"].Default)
		assert.Nil(t, flagsSchema.Properties["a_json_array"].Default)

		// verify json obj schemas
		parsedJSONObjSchema := flagsSchema.Properties["a_json_obj"]
		assert.Equal(t, aJSONObjSchema, parsedJSONObjSchema)

		// Verify array items schema
		includeSchema := flagsSchema.Properties["include"]
		assert.NotNil(t, includeSchema.Items)
		assert.Equal(t, "integer", includeSchema.Items.Type)
		greetingSchema := flagsSchema.Properties["greeting"]
		assert.NotNil(t, greetingSchema.Items)
		assert.Equal(t, "string", greetingSchema.Items.Type)

		// Verify stringToString object schema
		labelsSchema := flagsSchema.Properties["labels"]
		assert.NotNil(t, labelsSchema.AdditionalProperties)
		assert.Equal(t, "string", labelsSchema.AdditionalProperties.Type)

		// Verify stringToInt object schema
		portsSchema := flagsSchema.Properties["ports"]
		assert.NotNil(t, portsSchema.AdditionalProperties)
		assert.Equal(t, "integer", portsSchema.AdditionalProperties.Type)

		// Verify persistent flag from parent command
		assert.Contains(t, flagsSchema.Properties, "config", "Should include persistent flag from parent command")

		// Verify args schema
		argsSchema := inputSchema.Properties["args"]
		assert.Equal(t, "array", argsSchema.Type)
		assert.NotNil(t, argsSchema.Items)
		assert.Equal(t, "string", argsSchema.Items.Type)
	})

	t.Run("Restricted Selector", func(t *testing.T) {
		// Create a selector that only allows specific flags
		selector := Selector{
			LocalFlagSelector: func(flag *pflag.Flag) bool {
				names := []string{"output", "verbose", "hidden", "old"}
				return slices.Contains(names, flag.Name)
			},
			InheritedFlagSelector: func(_ *pflag.Flag) bool { return false },
		}

		// Create tool from command with the restricted selector
		tool := selector.createToolFromCmd(cmd, "parent")

		// Verify tool properties
		assert.Equal(t, "parent_test", tool.Name)
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

		// Verify excluded flags
		assert.NotContains(t, flagsSchema.Properties, "hidden", "Should not include hidden flag")
		assert.NotContains(t, flagsSchema.Properties, "old", "Should not include deprecated flag")
		assert.NotContains(t, flagsSchema.Properties, "include", "Should not include excluded flag")
		assert.NotContains(t, flagsSchema.Properties, "count", "Should not include excluded flag")
		assert.NotContains(t, flagsSchema.Properties, "config", "Should not include excluded persistent flag")
		assert.NotContains(t, flagsSchema.Properties, "greeting", "Should not include excluded flag")
		assert.NotContains(t, flagsSchema.Properties, "labels", "Should not include excluded flag")
		assert.NotContains(t, flagsSchema.Properties, "ports", "Should not include excluded flag")

		// Verify required flags - none should be required since 'count' was excluded
		require.Empty(t, flagsSchema.Required, "Should have no required flags")

		// Verify args schema
		argsSchema := inputSchema.Properties["args"]
		assert.Equal(t, "array", argsSchema.Type)
		assert.NotNil(t, argsSchema.Items)
		assert.Equal(t, "string", argsSchema.Items.Type)
	})
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

	t.Run("Default prefix (uses root name)", func(t *testing.T) {
		name := toolName(grandchild, "root")
		assert.Equal(t, "root_child_grandchild", name)
	})

	t.Run("Custom short prefix", func(t *testing.T) {
		name := toolName(grandchild, "r")
		assert.Equal(t, "r_child_grandchild", name)
	})

	t.Run("Root command only", func(t *testing.T) {
		name := toolName(root, "root")
		assert.Equal(t, "root", name)
	})

	t.Run("Root command with custom prefix", func(t *testing.T) {
		name := toolName(root, "myprefix")
		assert.Equal(t, "myprefix", name)
	})

	t.Run("Omnistrate use case - shortening long tool names", func(t *testing.T) {
		// Simulates: omnistrate-ctl cost by-instance-type in-provider
		omctl := &cobra.Command{Use: "omnistrate-ctl"}
		cost := &cobra.Command{Use: "cost"}
		byInstanceType := &cobra.Command{Use: "by-instance-type"}
		inProvider := &cobra.Command{Use: "in-provider", Run: func(_ *cobra.Command, _ []string) {}}

		omctl.AddCommand(cost)
		cost.AddCommand(byInstanceType)
		byInstanceType.AddCommand(inProvider)

		// Using full root name (original behavior)
		fullName := toolName(inProvider, "omnistrate-ctl")
		assert.Equal(t, "omnistrate-ctl_cost_by-instance-type_in-provider", fullName)

		// Using shortened prefix - saves 9 characters (len("omnistrate-ctl") - len("omctl") = 14 - 5 = 9)
		shortName := toolName(inProvider, "omctl")
		assert.Equal(t, "omctl_cost_by-instance-type_in-provider", shortName)
		assert.Less(t, len(shortName), len(fullName), "Short name should be shorter than full name")
		assert.Less(t, len(shortName), 64, "Short name should be under Claude's 64-char limit")
	})
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
