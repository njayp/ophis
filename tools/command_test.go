package tools

import (
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromRootCmd(t *testing.T) {
	tests := []struct {
		name     string
		setupCmd func() *cobra.Command
		expected int // number of tools expected
	}{
		{
			name: "simple root command with Run function",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{
					Use:   "root",
					Short: "Root command",
					Run:   func(_ *cobra.Command, _ []string) {},
				}
				return cmd
			},
			expected: 1,
		},
		{
			name: "simple root command with RunE function",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{
					Use:   "root",
					Short: "Root command",
					RunE:  func(_ *cobra.Command, _ []string) error { return nil },
				}
				return cmd
			},
			expected: 1,
		},
		{
			name: "root command without Run function",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{
					Use:   "root",
					Short: "Root command",
				}
				return cmd
			},
			expected: 0,
		},
		{
			name: "root command with subcommands",
			setupCmd: func() *cobra.Command {
				root := &cobra.Command{
					Use:   "root",
					Short: "Root command",
				}
				sub1 := &cobra.Command{
					Use:   "sub1",
					Short: "Subcommand 1",
					Run:   func(_ *cobra.Command, _ []string) {},
				}
				sub2 := &cobra.Command{
					Use:   "sub2",
					Short: "Subcommand 2",
					RunE:  func(_ *cobra.Command, _ []string) error { return nil },
				}
				root.AddCommand(sub1, sub2)
				return root
			},
			expected: 2,
		},
		{
			name: "root command with hidden subcommand",
			setupCmd: func() *cobra.Command {
				root := &cobra.Command{
					Use:   "root",
					Short: "Root command",
				}
				visible := &cobra.Command{
					Use:   "visible",
					Short: "Visible subcommand",
					Run:   func(_ *cobra.Command, _ []string) {},
				}
				hidden := &cobra.Command{
					Use:    "hidden",
					Short:  "Hidden subcommand",
					Hidden: true,
					Run:    func(_ *cobra.Command, _ []string) {},
				}
				root.AddCommand(visible, hidden)
				return root
			},
			expected: 1, // only visible command
		},
		{
			name: "root command with mcp subcommand (should be ignored)",
			setupCmd: func() *cobra.Command {
				root := &cobra.Command{
					Use:   "root",
					Short: "Root command",
				}
				mcpCmd := &cobra.Command{
					Use:   "mcp",
					Short: "MCP subcommand",
					Run:   func(_ *cobra.Command, _ []string) {},
				}
				other := &cobra.Command{
					Use:   "other",
					Short: "Other subcommand",
					Run:   func(_ *cobra.Command, _ []string) {},
				}
				root.AddCommand(mcpCmd, other)
				return root
			},
			expected: 1, // only 'other' command, 'mcp' ignored
		},
		{
			name: "nested subcommands",
			setupCmd: func() *cobra.Command {
				root := &cobra.Command{
					Use:   "root",
					Short: "Root command",
				}
				level1 := &cobra.Command{
					Use:   "level1",
					Short: "Level 1 command",
				}
				level2 := &cobra.Command{
					Use:   "level2",
					Short: "Level 2 command",
					Run:   func(_ *cobra.Command, _ []string) {},
				}
				level1.AddCommand(level2)
				root.AddCommand(level1)
				return root
			},
			expected: 1, // only level2 has Run function
		},
		{
			name: "deeply nested subcommands with multiple runnable commands",
			setupCmd: func() *cobra.Command {
				root := &cobra.Command{
					Use:   "root",
					Short: "Root command",
				}
				level1 := &cobra.Command{
					Use:   "level1",
					Short: "Level 1 command",
					Run:   func(_ *cobra.Command, _ []string) {},
				}
				level2a := &cobra.Command{
					Use:   "level2a",
					Short: "Level 2a command",
					Run:   func(_ *cobra.Command, _ []string) {},
				}
				level2b := &cobra.Command{
					Use:   "level2b",
					Short: "Level 2b command",
					RunE:  func(_ *cobra.Command, _ []string) error { return nil },
				}
				level3 := &cobra.Command{
					Use:   "level3",
					Short: "Level 3 command",
					Run:   func(_ *cobra.Command, _ []string) {},
				}
				level2a.AddCommand(level3)
				level1.AddCommand(level2a, level2b)
				root.AddCommand(level1)
				return root
			},
			expected: 4, // level1, level2a, level2b, level3
		},
		{
			name: "mixed runnable and non-runnable commands with hidden commands",
			setupCmd: func() *cobra.Command {
				root := &cobra.Command{
					Use:   "root",
					Short: "Root command",
					Run:   func(_ *cobra.Command, _ []string) {},
				}
				sub1 := &cobra.Command{
					Use:   "sub1",
					Short: "Subcommand 1",
					Run:   func(_ *cobra.Command, _ []string) {},
				}
				sub2 := &cobra.Command{
					Use:   "sub2",
					Short: "Subcommand 2",
					// No Run function
				}
				sub3 := &cobra.Command{
					Use:    "sub3",
					Short:  "Hidden subcommand 3",
					Hidden: true,
					Run:    func(_ *cobra.Command, _ []string) {},
				}
				sub4 := &cobra.Command{
					Use:   "sub4",
					Short: "Subcommand 4",
					RunE:  func(_ *cobra.Command, _ []string) error { return errors.New("test error") },
				}
				root.AddCommand(sub1, sub2, sub3, sub4)
				return root
			},
			expected: 3, // root, sub1, sub4 (sub2 has no Run, sub3 is hidden)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.setupCmd()
			tools := FromRootCmd(cmd)
			assert.Len(t, tools, tt.expected)

			// Verify each tool is properly constructed
			for _, tool := range tools {
				assert.NotEmpty(t, tool.Tool.Name)
				assert.NotNil(t, tool.Tool.InputSchema)
			}
		})
	}
}

func TestFlagToolOption(t *testing.T) {
	tests := []struct {
		name         string
		setupFlag    func() *cobra.Command
		flagName     string
		expectedType string
		checkDesc    bool
		expectedDesc string
	}{
		{
			name: "string flag",
			setupFlag: func() *cobra.Command {
				cmd := &cobra.Command{Use: "test"}
				cmd.Flags().String("test-flag", "", "Test string flag")
				return cmd
			},
			flagName:     "test-flag",
			expectedType: "string",
			checkDesc:    true,
			expectedDesc: "Test string flag",
		},
		{
			name: "boolean flag",
			setupFlag: func() *cobra.Command {
				cmd := &cobra.Command{Use: "test"}
				cmd.Flags().Bool("test-flag", false, "Test boolean flag")
				return cmd
			},
			flagName:     "test-flag",
			expectedType: "boolean",
			checkDesc:    true,
			expectedDesc: "Test boolean flag",
		},
		{
			name: "integer flag",
			setupFlag: func() *cobra.Command {
				cmd := &cobra.Command{Use: "test"}
				cmd.Flags().Int("test-flag", 0, "Test integer flag")
				return cmd
			},
			flagName:     "test-flag",
			expectedType: "integer",
			checkDesc:    true,
			expectedDesc: "Test integer flag",
		},
		{
			name: "int8 flag",
			setupFlag: func() *cobra.Command {
				cmd := &cobra.Command{Use: "test"}
				cmd.Flags().Int8("test-flag", 0, "Test int8 flag")
				return cmd
			},
			flagName:     "test-flag",
			expectedType: "integer",
			checkDesc:    true,
			expectedDesc: "Test int8 flag",
		},
		{
			name: "int16 flag",
			setupFlag: func() *cobra.Command {
				cmd := &cobra.Command{Use: "test"}
				cmd.Flags().Int16("test-flag", 0, "Test int16 flag")
				return cmd
			},
			flagName:     "test-flag",
			expectedType: "integer",
			checkDesc:    true,
			expectedDesc: "Test int16 flag",
		},
		{
			name: "int32 flag",
			setupFlag: func() *cobra.Command {
				cmd := &cobra.Command{Use: "test"}
				cmd.Flags().Int32("test-flag", 0, "Test int32 flag")
				return cmd
			},
			flagName:     "test-flag",
			expectedType: "integer",
			checkDesc:    true,
			expectedDesc: "Test int32 flag",
		},
		{
			name: "int64 flag",
			setupFlag: func() *cobra.Command {
				cmd := &cobra.Command{Use: "test"}
				cmd.Flags().Int64("test-flag", 0, "Test int64 flag")
				return cmd
			},
			flagName:     "test-flag",
			expectedType: "integer",
			checkDesc:    true,
			expectedDesc: "Test int64 flag",
		},
		{
			name: "uint flag",
			setupFlag: func() *cobra.Command {
				cmd := &cobra.Command{Use: "test"}
				cmd.Flags().Uint("test-flag", 0, "Test uint flag")
				return cmd
			},
			flagName:     "test-flag",
			expectedType: "integer",
			checkDesc:    true,
			expectedDesc: "Test uint flag",
		},
		{
			name: "uint8 flag",
			setupFlag: func() *cobra.Command {
				cmd := &cobra.Command{Use: "test"}
				cmd.Flags().Uint8("test-flag", 0, "Test uint8 flag")
				return cmd
			},
			flagName:     "test-flag",
			expectedType: "integer",
			checkDesc:    true,
			expectedDesc: "Test uint8 flag",
		},
		{
			name: "uint16 flag",
			setupFlag: func() *cobra.Command {
				cmd := &cobra.Command{Use: "test"}
				cmd.Flags().Uint16("test-flag", 0, "Test uint16 flag")
				return cmd
			},
			flagName:     "test-flag",
			expectedType: "integer",
			checkDesc:    true,
			expectedDesc: "Test uint16 flag",
		},
		{
			name: "uint32 flag",
			setupFlag: func() *cobra.Command {
				cmd := &cobra.Command{Use: "test"}
				cmd.Flags().Uint32("test-flag", 0, "Test uint32 flag")
				return cmd
			},
			flagName:     "test-flag",
			expectedType: "integer",
			checkDesc:    true,
			expectedDesc: "Test uint32 flag",
		},
		{
			name: "uint64 flag",
			setupFlag: func() *cobra.Command {
				cmd := &cobra.Command{Use: "test"}
				cmd.Flags().Uint64("test-flag", 0, "Test uint64 flag")
				return cmd
			},
			flagName:     "test-flag",
			expectedType: "integer",
			checkDesc:    true,
			expectedDesc: "Test uint64 flag",
		},
		{
			name: "float32 flag",
			setupFlag: func() *cobra.Command {
				cmd := &cobra.Command{Use: "test"}
				cmd.Flags().Float32("test-flag", 0.0, "Test float32 flag")
				return cmd
			},
			flagName:     "test-flag",
			expectedType: "number",
			checkDesc:    true,
			expectedDesc: "Test float32 flag",
		},
		{
			name: "float64 flag",
			setupFlag: func() *cobra.Command {
				cmd := &cobra.Command{Use: "test"}
				cmd.Flags().Float64("test-flag", 0.0, "Test float64 flag")
				return cmd
			},
			flagName:     "test-flag",
			expectedType: "number",
			checkDesc:    true,
			expectedDesc: "Test float64 flag",
		},
		{
			name: "string slice flag",
			setupFlag: func() *cobra.Command {
				cmd := &cobra.Command{Use: "test"}
				cmd.Flags().StringSlice("test-flag", nil, "Test string slice flag")
				return cmd
			},
			flagName:     "test-flag",
			expectedType: "stringArray",
			checkDesc:    true,
			expectedDesc: "Test string slice flag",
		},
		{
			name: "string array flag",
			setupFlag: func() *cobra.Command {
				cmd := &cobra.Command{Use: "test"}
				cmd.Flags().StringArray("test-flag", nil, "Test string array flag")
				return cmd
			},
			flagName:     "test-flag",
			expectedType: "stringArray",
			checkDesc:    true,
			expectedDesc: "Test string array flag",
		},
		{
			name: "int slice flag",
			setupFlag: func() *cobra.Command {
				cmd := &cobra.Command{Use: "test"}
				cmd.Flags().IntSlice("test-flag", nil, "Test int slice flag")
				return cmd
			},
			flagName:     "test-flag",
			expectedType: "intArray",
			checkDesc:    true,
			expectedDesc: "Test int slice flag",
		},
		{
			name: "flag with no usage description",
			setupFlag: func() *cobra.Command {
				cmd := &cobra.Command{Use: "test"}
				cmd.Flags().String("test-flag", "", "")
				return cmd
			},
			flagName:     "test-flag",
			expectedType: "string",
			checkDesc:    true,
			expectedDesc: "Flag: test-flag",
		},
		{
			name: "flag with whitespace-only description",
			setupFlag: func() *cobra.Command {
				cmd := &cobra.Command{Use: "test"}
				cmd.Flags().String("test-flag", "", "   ")
				return cmd
			},
			flagName:     "test-flag",
			expectedType: "string",
			checkDesc:    true,
			expectedDesc: "   ",
		},
		{
			name: "duration flag (should default to string)",
			setupFlag: func() *cobra.Command {
				cmd := &cobra.Command{Use: "test"}
				cmd.Flags().Duration("test-flag", 0, "Test duration flag")
				return cmd
			},
			flagName:     "test-flag",
			expectedType: "string",
			checkDesc:    true,
			expectedDesc: "Test duration flag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.setupFlag()
			var flag *pflag.Flag
			cmd.Flags().VisitAll(func(f *pflag.Flag) {
				if f.Name == tt.flagName {
					flag = f
				}
			})
			require.NotNil(t, flag, "Flag %s not found", tt.flagName)

			result := flagToolOption(flag)
			assert.Equal(t, tt.expectedType, result["type"])

			if tt.checkDesc {
				assert.Equal(t, tt.expectedDesc, result["description"])
			}

			// Verify result structure
			assert.Contains(t, result, "type")
			assert.Contains(t, result, "description")
			assert.Len(t, result, 2) // Should only have type and description
		})
	}
}

func TestFromRootCmd_ToolNaming(t *testing.T) {
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
				root := &cobra.Command{
					Use:   "root",
					Short: "Root command",
				}
				sub := &cobra.Command{
					Use:   "sub",
					Short: "Sub command",
					Run:   func(_ *cobra.Command, _ []string) {},
				}
				root.AddCommand(sub)
				return root
			},
			expectedName: "root_sub",
		},
		{
			name: "deeply nested command",
			setupCmd: func() *cobra.Command {
				root := &cobra.Command{
					Use:   "root",
					Short: "Root command",
				}
				level1 := &cobra.Command{
					Use:   "level1",
					Short: "Level 1 command",
				}
				level2 := &cobra.Command{
					Use:   "level2",
					Short: "Level 2 command",
					Run:   func(_ *cobra.Command, _ []string) {},
				}
				level1.AddCommand(level2)
				root.AddCommand(level1)
				return root
			},
			expectedName: "root_level1_level2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.setupCmd()
			tools := FromRootCmd(cmd)
			require.Len(t, tools, 1)
			assert.Equal(t, tt.expectedName, tools[0].Tool.Name)
		})
	}
}

func TestArgsDescFromCmd(t *testing.T) {
	tests := []struct {
		name     string
		setupCmd func() *cobra.Command
		expected string
	}{
		{
			name: "command with Use field",
			setupCmd: func() *cobra.Command {
				return &cobra.Command{
					Use: "test [OPTIONS] <arg1> <arg2>",
				}
			},
			expected: "Positional arguments. Usage: test [OPTIONS] <arg1> <arg2>",
		},
		{
			name: "command without Use field",
			setupCmd: func() *cobra.Command {
				return &cobra.Command{
					Short: "Test command",
				}
			},
			expected: "Positional arguments",
		},
		{
			name: "command with empty Use field",
			setupCmd: func() *cobra.Command {
				return &cobra.Command{
					Use:   "",
					Short: "Test command",
				}
			},
			expected: "Positional arguments",
		},
		{
			name: "command with complex Use syntax",
			setupCmd: func() *cobra.Command {
				return &cobra.Command{
					Use: "kubectl get [(-o|--output=)json|yaml|name|go-template|go-template-file|template|templatefile|jsonpath|jsonpath-as-json|jsonpath-file|custom-columns|custom-columns-file|wide] (TYPE[.VERSION][.GROUP] [NAME | -l label] | TYPE[.VERSION][.GROUP]/NAME ...) [flags]",
				}
			},
			expected: "Positional arguments. Usage: kubectl get [(-o|--output=)json|yaml|name|go-template|go-template-file|template|templatefile|jsonpath|jsonpath-as-json|jsonpath-file|custom-columns|custom-columns-file|wide] (TYPE[.VERSION][.GROUP] [NAME | -l label] | TYPE[.VERSION][.GROUP]/NAME ...) [flags]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.setupCmd()
			result := argsDescFromCmd(cmd)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDescFromCmd(t *testing.T) {
	tests := []struct {
		name     string
		setupCmd func() *cobra.Command
		expected string
	}{
		{
			name: "command with long description",
			setupCmd: func() *cobra.Command {
				return &cobra.Command{
					Short: "Short description",
					Long:  "Long description with more details",
				}
			},
			expected: "Long description with more details",
		},
		{
			name: "command with only short description",
			setupCmd: func() *cobra.Command {
				return &cobra.Command{
					Short: "Short description",
				}
			},
			expected: "Short description",
		},
		{
			name: "command with empty descriptions",
			setupCmd: func() *cobra.Command {
				return &cobra.Command{
					Use: "test",
				}
			},
			expected: "",
		},
		{
			name: "command with empty long but has short",
			setupCmd: func() *cobra.Command {
				return &cobra.Command{
					Short: "Short description",
					Long:  "",
				}
			},
			expected: "Short description",
		},
		{
			name: "command with whitespace-only long description",
			setupCmd: func() *cobra.Command {
				return &cobra.Command{
					Short: "Short description",
					Long:  "   ",
				}
			},
			expected: "   ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.setupCmd()
			result := descFromCmd(cmd)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFlagMapFromCmd(t *testing.T) {
	tests := []struct {
		name     string
		setupCmd func(t *testing.T) *cobra.Command
		validate func(t *testing.T, flagMap map[string]any)
	}{
		{
			name: "command with no flags",
			setupCmd: func(_ *testing.T) *cobra.Command {
				return &cobra.Command{
					Use: "test",
					Run: func(_ *cobra.Command, _ []string) {},
				}
			},
			validate: func(t *testing.T, flagMap map[string]any) {
				assert.Empty(t, flagMap)
			},
		},
		{
			name: "command with local flags",
			setupCmd: func(_ *testing.T) *cobra.Command {
				cmd := &cobra.Command{
					Use: "test",
					Run: func(_ *cobra.Command, _ []string) {},
				}
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
				assert.Equal(t, "A string flag", stringFlag["description"])

				intFlag := flagMap["int-flag"].(map[string]string)
				assert.Equal(t, "integer", intFlag["type"])
				assert.Equal(t, "An integer flag", intFlag["description"])

				boolFlag := flagMap["bool-flag"].(map[string]string)
				assert.Equal(t, "boolean", boolFlag["type"])
				assert.Equal(t, "A boolean flag", boolFlag["description"])
			},
		},
		{
			name: "command with hidden flags",
			setupCmd: func(_ *testing.T) *cobra.Command {
				cmd := &cobra.Command{
					Use: "test",
					Run: func(_ *cobra.Command, _ []string) {},
				}
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
			setupCmd: func(_ *testing.T) *cobra.Command {
				parent := &cobra.Command{
					Use: "parent",
				}
				parent.PersistentFlags().String("persistent-flag", "", "A persistent flag")

				child := &cobra.Command{
					Use: "child",
					Run: func(_ *cobra.Command, _ []string) {},
				}
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
		{
			name: "command with overlapping local and persistent flags",
			setupCmd: func(_ *testing.T) *cobra.Command {
				parent := &cobra.Command{
					Use: "parent",
				}
				parent.PersistentFlags().String("common-flag", "", "A persistent flag")
				parent.PersistentFlags().String("persistent-only", "", "Persistent only flag")

				child := &cobra.Command{
					Use: "child",
					Run: func(_ *cobra.Command, _ []string) {},
				}
				child.Flags().String("common-flag", "", "A local flag that overrides persistent")
				child.Flags().String("local-only", "", "Local only flag")
				parent.AddCommand(child)
				return child
			},
			validate: func(t *testing.T, flagMap map[string]any) {
				assert.Len(t, flagMap, 3)
				assert.Contains(t, flagMap, "common-flag")
				assert.Contains(t, flagMap, "local-only")
				assert.Contains(t, flagMap, "persistent-only")

				// Verify that local flag takes precedence over persistent
				commonFlag := flagMap["common-flag"].(map[string]string)
				assert.Equal(t, "A local flag that overrides persistent", commonFlag["description"])
			},
		},
		{
			name: "command with various flag types",
			setupCmd: func(_ *testing.T) *cobra.Command {
				cmd := &cobra.Command{
					Use: "test",
					Run: func(_ *cobra.Command, _ []string) {},
				}
				cmd.Flags().String("string-flag", "", "String flag")
				cmd.Flags().Int("int-flag", 0, "Int flag")
				cmd.Flags().Int64("int64-flag", 0, "Int64 flag")
				cmd.Flags().Float32("float32-flag", 0.0, "Float32 flag")
				cmd.Flags().Float64("float64-flag", 0.0, "Float64 flag")
				cmd.Flags().Bool("bool-flag", false, "Bool flag")
				cmd.Flags().StringSlice("stringslice-flag", nil, "StringSlice flag")
				cmd.Flags().StringArray("stringarray-flag", nil, "StringArray flag")
				cmd.Flags().IntSlice("intslice-flag", nil, "IntSlice flag")
				return cmd
			},
			validate: func(t *testing.T, flagMap map[string]any) {
				assert.Len(t, flagMap, 9)

				// Test string types
				stringFlag := flagMap["string-flag"].(map[string]string)
				assert.Equal(t, "string", stringFlag["type"])

				// Test integer types
				intFlag := flagMap["int-flag"].(map[string]string)
				assert.Equal(t, "integer", intFlag["type"])
				int64Flag := flagMap["int64-flag"].(map[string]string)
				assert.Equal(t, "integer", int64Flag["type"])

				// Test number types
				float32Flag := flagMap["float32-flag"].(map[string]string)
				assert.Equal(t, "number", float32Flag["type"])
				float64Flag := flagMap["float64-flag"].(map[string]string)
				assert.Equal(t, "number", float64Flag["type"])

				// Test boolean type
				boolFlag := flagMap["bool-flag"].(map[string]string)
				assert.Equal(t, "boolean", boolFlag["type"])

				// Test array types
				stringSliceFlag := flagMap["stringslice-flag"].(map[string]string)
				assert.Equal(t, "stringArray", stringSliceFlag["type"])
				stringArrayFlag := flagMap["stringarray-flag"].(map[string]string)
				assert.Equal(t, "stringArray", stringArrayFlag["type"])
				intSliceFlag := flagMap["intslice-flag"].(map[string]string)
				assert.Equal(t, "intArray", intSliceFlag["type"])
			},
		},
		{
			name: "command with hidden persistent flags",
			setupCmd: func(t *testing.T) *cobra.Command {
				parent := &cobra.Command{
					Use: "parent",
				}
				parent.PersistentFlags().String("visible-persistent", "", "Visible persistent flag")
				parent.PersistentFlags().String("hidden-persistent", "", "Hidden persistent flag")
				err := parent.PersistentFlags().MarkHidden("hidden-persistent")
				require.NoError(t, err)

				child := &cobra.Command{
					Use: "child",
					Run: func(_ *cobra.Command, _ []string) {},
				}
				child.Flags().String("local-flag", "", "Local flag")
				parent.AddCommand(child)
				return child
			},
			validate: func(t *testing.T, flagMap map[string]any) {
				assert.Len(t, flagMap, 2)
				assert.Contains(t, flagMap, "visible-persistent")
				assert.Contains(t, flagMap, "local-flag")
				assert.NotContains(t, flagMap, "hidden-persistent")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.setupCmd(t)
			flagMap := flagMapFromCmd(cmd)
			tt.validate(t, flagMap)
		})
	}
}
