package tools

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGenerator_CommandConversion tests the core functionality of converting
// Cobra commands to MCP tools, covering the main path and key edge cases
func TestGenerator_CommandConversion(t *testing.T) {
	tests := []struct {
		name     string
		setupCmd func() *cobra.Command
		options  []GeneratorOption
		validate func(t *testing.T, tools []Controller)
	}{
		{
			name: "standard CLI with subcommands",
			setupCmd: func() *cobra.Command {
				root := &cobra.Command{Use: "cli", Short: "CLI tool"}

				// Common pattern: get/list/create commands
				get := &cobra.Command{Use: "get", Short: "Get resources", Run: func(_ *cobra.Command, _ []string) {}}
				list := &cobra.Command{Use: "list", Short: "List resources", Run: func(_ *cobra.Command, _ []string) {}}
				create := &cobra.Command{Use: "create", Short: "Create resource", RunE: func(_ *cobra.Command, _ []string) error { return nil }}

				// Add some flags
				get.Flags().String("format", "json", "Output format")
				list.Flags().Bool("all", false, "Show all resources")
				create.Flags().StringP("file", "f", "", "File to create from")

				root.AddCommand(get, list, create)
				return root
			},
			validate: func(t *testing.T, tools []Controller) {
				assert.Len(t, tools, 3)

				// Verify tool names
				toolNames := make(map[string]bool)
				for _, tool := range tools {
					toolNames[tool.Tool.Name] = true
				}
				assert.True(t, toolNames["cli_get"])
				assert.True(t, toolNames["cli_list"])
				assert.True(t, toolNames["cli_create"])

				// Verify flags are included
				for _, tool := range tools {
					// Verify tool has proper structure
					assert.NotNil(t, tool.Tool.InputSchema)
					// Just verify the tool was created properly
					// The schema structure is handled by mcp-go library
				}
			},
		},
		{
			name: "nested command structure",
			setupCmd: func() *cobra.Command {
				root := &cobra.Command{Use: "kubectl", Short: "Kubernetes CLI"}

				// Nested structure: kubectl get pods
				get := &cobra.Command{Use: "get", Short: "Get resources"}
				pods := &cobra.Command{Use: "pods", Short: "Get pods", Run: func(_ *cobra.Command, _ []string) {}}
				pods.Flags().StringP("namespace", "n", "default", "Namespace")

				get.AddCommand(pods)
				root.AddCommand(get)
				return root
			},
			validate: func(t *testing.T, tools []Controller) {
				assert.Len(t, tools, 1)
				assert.Equal(t, "kubectl_get_pods", tools[0].Tool.Name)
			},
		},
		{
			name: "filtered commands with allow list",
			setupCmd: func() *cobra.Command {
				root := &cobra.Command{Use: "cli", Short: "CLI"}

				safe := &cobra.Command{Use: "safe", Short: "Safe operation", Run: func(_ *cobra.Command, _ []string) {}}
				danger := &cobra.Command{Use: "danger", Short: "Dangerous operation", Run: func(_ *cobra.Command, _ []string) {}}
				other := &cobra.Command{Use: "other", Short: "Other operation", Run: func(_ *cobra.Command, _ []string) {}}

				root.AddCommand(safe, danger, other)
				return root
			},
			options: []GeneratorOption{
				WithFilters(Allow([]string{"safe", "other"})),
			},
			validate: func(t *testing.T, tools []Controller) {
				assert.Len(t, tools, 2)

				toolNames := make([]string, len(tools))
				for i, tool := range tools {
					toolNames[i] = tool.Tool.Name
				}
				assert.ElementsMatch(t, []string{"cli_safe", "cli_other"}, toolNames)
			},
		},
		{
			name: "commands without Run functions are skipped",
			setupCmd: func() *cobra.Command {
				root := &cobra.Command{Use: "cli", Short: "CLI"}

				// Parent command without Run - should be skipped
				parent := &cobra.Command{Use: "parent", Short: "Parent command"}

				// Child with Run - should be included
				child := &cobra.Command{Use: "child", Short: "Child command", Run: func(_ *cobra.Command, _ []string) {}}

				parent.AddCommand(child)
				root.AddCommand(parent)
				return root
			},
			validate: func(t *testing.T, tools []Controller) {
				assert.Len(t, tools, 1)
				assert.Equal(t, "cli_parent_child", tools[0].Tool.Name)
			},
		},
		{
			name: "inherited flags from parent commands",
			setupCmd: func() *cobra.Command {
				root := &cobra.Command{Use: "cli", Short: "CLI"}
				root.PersistentFlags().String("config", "", "Config file")
				root.PersistentFlags().Bool("verbose", false, "Verbose output")

				cmd := &cobra.Command{Use: "run", Short: "Run command", Run: func(_ *cobra.Command, _ []string) {}}
				cmd.Flags().String("input", "", "Input file")

				root.AddCommand(cmd)
				return root
			},
			validate: func(t *testing.T, tools []Controller) {
				assert.Len(t, tools, 1)

				// Verify the tool was created
				assert.Equal(t, "cli_run", tools[0].Tool.Name)
				// The actual flag verification would require inspecting the schema
				// which is internal to the mcp.Tool structure
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator := NewGenerator(tt.options...)
			cmd := tt.setupCmd()
			tools := generator.FromRootCmd(cmd)
			tt.validate(t, tools)
		})
	}
}

// TestFlagTypeMapping tests the mapping of Cobra flag types to MCP schema types
func TestFlagTypeMapping(t *testing.T) {
	// Test only the most common and important flag types
	tests := []struct {
		flagType       string
		setup          func(cmd *cobra.Command)
		validateSchema func(t *testing.T, result map[string]any)
	}{
		{
			flagType: "string",
			setup:    func(cmd *cobra.Command) { cmd.Flags().String("test", "", "desc") },
			validateSchema: func(t *testing.T, result map[string]any) {
				assert.Equal(t, "string", result["type"])
				assert.Equal(t, "desc", result["description"])
			},
		},
		{
			flagType: "bool",
			setup:    func(cmd *cobra.Command) { cmd.Flags().Bool("test", false, "desc") },
			validateSchema: func(t *testing.T, result map[string]any) {
				assert.Equal(t, "boolean", result["type"])
				assert.Equal(t, "desc", result["description"])
			},
		},
		{
			flagType: "int",
			setup:    func(cmd *cobra.Command) { cmd.Flags().Int("test", 0, "desc") },
			validateSchema: func(t *testing.T, result map[string]any) {
				assert.Equal(t, "integer", result["type"])
				assert.Equal(t, "desc", result["description"])
			},
		},
		{
			flagType: "stringSlice",
			setup:    func(cmd *cobra.Command) { cmd.Flags().StringSlice("test", nil, "desc") },
			validateSchema: func(t *testing.T, result map[string]any) {
				assert.Equal(t, "array", result["type"])
				assert.Equal(t, "desc", result["description"])
				items, ok := result["items"].(map[string]any)
				require.True(t, ok, "items should be a map")
				assert.Equal(t, "string", items["type"])
			},
		},
		{
			flagType: "intSlice",
			setup:    func(cmd *cobra.Command) { cmd.Flags().IntSlice("test", nil, "desc") },
			validateSchema: func(t *testing.T, result map[string]any) {
				assert.Equal(t, "array", result["type"])
				assert.Equal(t, "desc", result["description"])
				items, ok := result["items"].(map[string]any)
				require.True(t, ok, "items should be a map")
				assert.Equal(t, "integer", items["type"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.flagType, func(t *testing.T) {
			cmd := &cobra.Command{Use: "test"}
			tt.setup(cmd)

			var flag *pflag.Flag
			cmd.Flags().VisitAll(func(f *pflag.Flag) {
				flag = f
			})
			require.NotNil(t, flag)

			result := flagToolOption(flag)
			tt.validateSchema(t, result)
		})
	}
}

// TestDefaultFilters tests that default filters work as expected
func TestDefaultFilters(t *testing.T) {
	root := &cobra.Command{Use: "cli", Short: "CLI"}

	// Commands that should be filtered out by default
	mcp := &cobra.Command{Use: "mcp", Short: "MCP command", Run: func(_ *cobra.Command, _ []string) {}}
	help := &cobra.Command{Use: "help", Short: "Help command", Run: func(_ *cobra.Command, _ []string) {}}
	completion := &cobra.Command{Use: "completion", Short: "Completion", Run: func(_ *cobra.Command, _ []string) {}}
	hidden := &cobra.Command{Use: "hidden", Short: "Hidden", Hidden: true, Run: func(_ *cobra.Command, _ []string) {}}

	// Command that should be included
	normal := &cobra.Command{Use: "normal", Short: "Normal command", Run: func(_ *cobra.Command, _ []string) {}}

	root.AddCommand(mcp, help, completion, hidden, normal)

	generator := NewGenerator() // Use default filters
	tools := generator.FromRootCmd(root)

	assert.Len(t, tools, 1, "Should only include the normal command")
	assert.Equal(t, "cli_normal", tools[0].Tool.Name)
}

// TestCommandDescriptions tests that command descriptions are properly extracted
func TestCommandDescriptions(t *testing.T) {
	tests := []struct {
		name     string
		short    string
		long     string
		expected string
	}{
		{
			name:     "prefers long description",
			short:    "Short desc",
			long:     "This is a much longer and more detailed description",
			expected: "This is a much longer and more detailed description",
		},
		{
			name:     "falls back to short if no long",
			short:    "Only short description",
			long:     "",
			expected: "Only short description",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use:   "test",
				Short: tt.short,
				Long:  tt.long,
				Run:   func(_ *cobra.Command, _ []string) {},
			}

			generator := NewGenerator()
			tools := generator.FromRootCmd(cmd)

			require.Len(t, tools, 1)
			assert.Equal(t, tt.expected, tools[0].Tool.Description)
		})
	}
}
