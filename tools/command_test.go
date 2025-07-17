package tools

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerator_FromRootCmd(t *testing.T) {
	tests := []struct {
		name     string
		setupCmd func() *cobra.Command
		expected int // number of tools expected
	}{
		{
			name: "root command with Run function",
			setupCmd: func() *cobra.Command {
				return &cobra.Command{
					Use:   "root",
					Short: "Root command",
					Run:   func(_ *cobra.Command, _ []string) {},
				}
			},
			expected: 1,
		},
		{
			name: "root command without Run function",
			setupCmd: func() *cobra.Command {
				return &cobra.Command{
					Use:   "root",
					Short: "Root command",
				}
			},
			expected: 0,
		},
		{
			name: "root with subcommands",
			setupCmd: func() *cobra.Command {
				root := &cobra.Command{Use: "root", Short: "Root command"}
				sub1 := &cobra.Command{Use: "sub1", Short: "Sub 1", Run: func(_ *cobra.Command, _ []string) {}}
				sub2 := &cobra.Command{Use: "sub2", Short: "Sub 2", RunE: func(_ *cobra.Command, _ []string) error { return nil }}
				root.AddCommand(sub1, sub2)
				return root
			},
			expected: 2,
		},
		{
			name: "root with hidden subcommand",
			setupCmd: func() *cobra.Command {
				root := &cobra.Command{Use: "root", Short: "Root command"}
				visible := &cobra.Command{Use: "visible", Short: "Visible", Run: func(_ *cobra.Command, _ []string) {}}
				hidden := &cobra.Command{Use: "hidden", Short: "Hidden", Hidden: true, Run: func(_ *cobra.Command, _ []string) {}}
				root.AddCommand(visible, hidden)
				return root
			},
			expected: 1, // only visible command
		},
		{
			name: "root with mcp subcommand (excluded by default)",
			setupCmd: func() *cobra.Command {
				root := &cobra.Command{Use: "root", Short: "Root command"}
				mcpCmd := &cobra.Command{Use: "mcp", Short: "MCP", Run: func(_ *cobra.Command, _ []string) {}}
				other := &cobra.Command{Use: "other", Short: "Other", Run: func(_ *cobra.Command, _ []string) {}}
				root.AddCommand(mcpCmd, other)
				return root
			},
			expected: 1, // only 'other' command, 'mcp' excluded
		},
		{
			name: "command tree with multiple branches",
			setupCmd: func() *cobra.Command {
				root := &cobra.Command{Use: "root", Short: "Root command"}
				// branch 1
				cmd11 := &cobra.Command{Use: "cmd11", Short: "11", Run: func(_ *cobra.Command, _ []string) {}}
				cmd12 := &cobra.Command{Use: "cmd12", Short: "12", Run: func(_ *cobra.Command, _ []string) {}}
				cmd11.AddCommand(cmd12)
				// branch 2
				cmd21 := &cobra.Command{Use: "cmd21", Short: "21", Run: func(_ *cobra.Command, _ []string) {}}
				cmd22 := &cobra.Command{Use: "cmd22", Short: "22", Run: func(_ *cobra.Command, _ []string) {}}
				cmd21.AddCommand(cmd22)
				// Add branches to root
				root.AddCommand(cmd11, cmd21)
				return root
			},
			expected: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator := NewGenerator()
			cmd := tt.setupCmd()
			tools := generator.FromRootCmd(cmd)

			assert.Len(t, tools, tt.expected)

			// Verify each tool is properly constructed
			for _, tool := range tools {
				assert.NotEmpty(t, tool.Tool.Name)
				assert.NotNil(t, tool.Tool.InputSchema)
			}
		})
	}
}

func TestGenerator_BlackWhiteListOptions(t *testing.T) {
	tests := []struct {
		name     string
		options  []GeneratorOption
		expected []string
	}{
		{
			name:     "blacklist custom",
			options:  []GeneratorOption{WithFilters(Exclude([]string{"custom"}))},
			expected: []string{"root_other"},
		},
		{
			name:     "no list (default: blacklist mcp)",
			options:  nil,
			expected: []string{"root_custom", "root_other", "root_custom_child"},
		},
		{
			name:     "whitelist custom",
			options:  []GeneratorOption{WithFilters(Allow([]string{"custom"}))},
			expected: []string{"root_custom", "root_custom_child"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := &cobra.Command{Use: "root", Short: "Root command"}
			customCmd := &cobra.Command{Use: "custom", Short: "Custom", Run: func(_ *cobra.Command, _ []string) {}}
			childCmd := &cobra.Command{Use: "child", Short: "Child", Run: func(_ *cobra.Command, _ []string) {}}
			otherCmd := &cobra.Command{Use: "other", Short: "Other", Run: func(_ *cobra.Command, _ []string) {}}
			root.AddCommand(customCmd, otherCmd)
			customCmd.AddCommand(childCmd)

			generator := NewGenerator(tt.options...)
			tools := generator.FromRootCmd(root)

			var toolNames []string
			for _, tool := range tools {
				toolNames = append(toolNames, tool.Tool.Name)
			}
			assert.ElementsMatch(t, tt.expected, toolNames)
		})
	}
}

func TestFlagToolOption(t *testing.T) {
	tests := []struct {
		name         string
		flagType     string
		expectedType string
	}{
		{"string flag", "string", "string"},
		{"boolean flag", "bool", "boolean"},
		{"integer flag", "int", "integer"},
		{"int64 flag", "int64", "integer"},
		{"float32 flag", "float32", "number"},
		{"float64 flag", "float64", "number"},
		{"string slice flag", "stringSlice", "stringArray"},
		{"string array flag", "stringArray", "stringArray"},
		{"int slice flag", "intSlice", "intArray"},
		{"duration flag (default to string)", "duration", "string"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "test"}

			// Add different flag types based on test case
			switch tt.flagType {
			case "string":
				cmd.Flags().String("test-flag", "", "Test flag")
			case "bool":
				cmd.Flags().Bool("test-flag", false, "Test flag")
			case "int":
				cmd.Flags().Int("test-flag", 0, "Test flag")
			case "int64":
				cmd.Flags().Int64("test-flag", 0, "Test flag")
			case "float32":
				cmd.Flags().Float32("test-flag", 0.0, "Test flag")
			case "float64":
				cmd.Flags().Float64("test-flag", 0.0, "Test flag")
			case "stringSlice":
				cmd.Flags().StringSlice("test-flag", nil, "Test flag")
			case "stringArray":
				cmd.Flags().StringArray("test-flag", nil, "Test flag")
			case "intSlice":
				cmd.Flags().IntSlice("test-flag", nil, "Test flag")
			case "duration":
				cmd.Flags().Duration("test-flag", 0, "Test flag")
			}

			var flag *pflag.Flag
			cmd.Flags().VisitAll(func(f *pflag.Flag) {
				if f.Name == "test-flag" {
					flag = f
				}
			})
			require.NotNil(t, flag, "Flag test-flag not found")

			result := flagToolOption(flag)
			assert.Equal(t, tt.expectedType, result["type"])
			assert.Equal(t, "Test flag", result["description"])
			assert.Len(t, result, 2) // Should only have type and description
		})
	}
}

func TestToolNaming(t *testing.T) {
	tests := []struct {
		name         string
		setupCmd     func() *cobra.Command
		expectedName string
	}{
		{
			name: "single level command",
			setupCmd: func() *cobra.Command {
				return &cobra.Command{
					Use:   "mycommand",
					Short: "My command",
					Run:   func(_ *cobra.Command, _ []string) {},
				}
			},
			expectedName: "mycommand",
		},
		{
			name: "nested command",
			setupCmd: func() *cobra.Command {
				root := &cobra.Command{Use: "root", Short: "Root command"}
				sub := &cobra.Command{Use: "sub", Short: "Sub command", Run: func(_ *cobra.Command, _ []string) {}}
				root.AddCommand(sub)
				return root
			},
			expectedName: "root_sub",
		},
		{
			name: "deeply nested command",
			setupCmd: func() *cobra.Command {
				root := &cobra.Command{Use: "root", Short: "Root command"}
				level1 := &cobra.Command{Use: "level1", Short: "Level 1 command"}
				level2 := &cobra.Command{Use: "level2", Short: "Level 2 command", Run: func(_ *cobra.Command, _ []string) {}}
				level1.AddCommand(level2)
				root.AddCommand(level1)
				return root
			},
			expectedName: "root_level1_level2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator := NewGenerator()
			cmd := tt.setupCmd()
			tools := generator.FromRootCmd(cmd)
			require.Len(t, tools, 1)
			assert.Equal(t, tt.expectedName, tools[0].Tool.Name)
		})
	}
}

func TestArgsDescFromCmd(t *testing.T) {
	tests := []struct {
		name     string
		use      string
		expected string
	}{
		{
			name:     "command with Use field",
			use:      "test [OPTIONS] <arg1> <arg2>",
			expected: "Positional arguments. Usage: test [OPTIONS] <arg1> <arg2>",
		},
		{
			name:     "command without Use field",
			use:      "",
			expected: "Positional arguments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: tt.use}
			result := argsDescFromCmd(cmd)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDescFromCmd(t *testing.T) {
	tests := []struct {
		name     string
		short    string
		long     string
		expected string
	}{
		{
			name:     "command with long description",
			short:    "Short description",
			long:     "Long description with more details",
			expected: "Long description with more details",
		},
		{
			name:     "command with only short description",
			short:    "Short description",
			long:     "",
			expected: "Short description",
		},
		{
			name:     "command with empty descriptions",
			short:    "",
			long:     "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{Short: tt.short, Long: tt.long}
			result := descFromCmd(cmd)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFlagMapFromCmd(t *testing.T) {
	tests := []struct {
		name     string
		setupCmd func() *cobra.Command
		validate func(t *testing.T, flagMap map[string]any)
	}{
		{
			name: "command with no flags",
			setupCmd: func() *cobra.Command {
				return &cobra.Command{Use: "test", Run: func(_ *cobra.Command, _ []string) {}}
			},
			validate: func(t *testing.T, flagMap map[string]any) {
				assert.Empty(t, flagMap)
			},
		},
		{
			name: "command with basic flags",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{Use: "test", Run: func(_ *cobra.Command, _ []string) {}}
				cmd.Flags().String("string-flag", "", "A string flag")
				cmd.Flags().Int("int-flag", 0, "An integer flag")
				cmd.Flags().Bool("bool-flag", false, "A boolean flag")
				return cmd
			},
			validate: func(t *testing.T, flagMap map[string]any) {
				assert.Len(t, flagMap, 3)
				assert.Contains(t, flagMap, "string-flag")
				assert.Contains(t, flagMap, "int-flag")
				assert.Contains(t, flagMap, "bool-flag")

				// Verify flag types
				stringFlag := flagMap["string-flag"].(map[string]string)
				assert.Equal(t, "string", stringFlag["type"])

				intFlag := flagMap["int-flag"].(map[string]string)
				assert.Equal(t, "integer", intFlag["type"])

				boolFlag := flagMap["bool-flag"].(map[string]string)
				assert.Equal(t, "boolean", boolFlag["type"])
			},
		},
		{
			name: "command with hidden flags",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{Use: "test", Run: func(_ *cobra.Command, _ []string) {}}
				cmd.Flags().String("visible-flag", "", "A visible flag")
				cmd.Flags().String("hidden-flag", "", "A hidden flag")
				err := cmd.Flags().MarkHidden("hidden-flag")
				require.NoError(t, err)
				return cmd
			},
			validate: func(t *testing.T, flagMap map[string]any) {
				assert.Len(t, flagMap, 1)
				assert.Contains(t, flagMap, "visible-flag")
				assert.NotContains(t, flagMap, "hidden-flag")
			},
		},
		{
			name: "command with persistent flags",
			setupCmd: func() *cobra.Command {
				parent := &cobra.Command{Use: "parent"}
				parent.PersistentFlags().String("persistent-flag", "", "A persistent flag")

				child := &cobra.Command{Use: "child", Run: func(_ *cobra.Command, _ []string) {}}
				child.Flags().String("local-flag", "", "A local flag")
				parent.AddCommand(child)
				return child
			},
			validate: func(t *testing.T, flagMap map[string]any) {
				assert.Len(t, flagMap, 2)
				assert.Contains(t, flagMap, "local-flag")
				assert.Contains(t, flagMap, "persistent-flag")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.setupCmd()
			flagMap := flagMapFromCmd(cmd)
			tt.validate(t, flagMap)
		})
	}
}
