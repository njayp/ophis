package ophis

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/spf13/cobra"
)

// mockCallToolRequest implements the mcp.CallToolRequest interface for testing
type mockCallToolRequest struct {
	Params struct {
		Arguments map[string]interface{}
	}
}

func (m *mockCallToolRequest) GetArguments() map[string]interface{} {
	return m.Params.Arguments
}

// TestCobraToMCPBridge_executeCommand tests the executeCommand method with various message scenarios
func TestCobraToMCPBridge_executeCommand(t *testing.T) {
	tests := []struct {
		name           string
		setupCmd       func() *cobra.Command
		parentPath     string
		args           map[string]interface{}
		expectedOutput string
		expectError    bool
		description    string
	}{
		{
			name: "simple hello command with no args",
			setupCmd: func() *cobra.Command {
				return &cobra.Command{
					Use:   "hello",
					Short: "Say hello",
					Run: func(cmd *cobra.Command, args []string) {
						cmd.Print("Hello, World!")
					},
				}
			},
			parentPath:     "",
			args:           map[string]interface{}{},
			expectedOutput: "Hello, World!",
			expectError:    false,
			description:    "Basic command with no arguments should execute successfully",
		},
		{
			name: "hello command with positional argument",
			setupCmd: func() *cobra.Command {
				return &cobra.Command{
					Use:   "hello [name]",
					Short: "Say hello to someone",
					Run: func(cmd *cobra.Command, args []string) {
						name := "World"
						if len(args) > 0 {
							name = args[0]
						}
						cmd.Printf("Hello, %s!", name)
					},
				}
			},
			parentPath: "",
			args: map[string]interface{}{
				"args": "Alice",
			},
			expectedOutput: "Hello, Alice!",
			expectError:    false,
			description:    "Command with positional argument should pass the argument correctly",
		},
		{
			name: "hello command with multiple positional arguments",
			setupCmd: func() *cobra.Command {
				return &cobra.Command{
					Use:   "greet [names...]",
					Short: "Greet multiple people",
					Run: func(cmd *cobra.Command, args []string) {
						if len(args) == 0 {
							cmd.Print("Hello, everyone!")
						} else {
							cmd.Printf("Hello, %s!", strings.Join(args, ", "))
						}
					},
				}
			},
			parentPath: "",
			args: map[string]interface{}{
				"args": "Alice Bob Charlie",
			},
			expectedOutput: "Hello, Alice, Bob, Charlie!",
			expectError:    false,
			description:    "Command with multiple positional arguments should split and pass them correctly",
		},
		{
			name: "command with string flag",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{
					Use:   "hello",
					Short: "Say hello with custom greeting",
					Run: func(cmd *cobra.Command, args []string) {
						greeting, _ := cmd.Flags().GetString("greeting")
						name := "World"
						if len(args) > 0 {
							name = args[0]
						}
						cmd.Printf("%s, %s!", greeting, name)
					},
				}
				cmd.Flags().String("greeting", "Hello", "The greeting to use")
				return cmd
			},
			parentPath: "",
			args: map[string]interface{}{
				"greeting": "Hi",
				"args":     "Bob",
			},
			expectedOutput: "Hi, Bob!",
			expectError:    false,
			description:    "Command with string flag should set the flag value correctly",
		},
		{
			name: "command with boolean flag",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{
					Use:   "list",
					Short: "List items",
					Run: func(cmd *cobra.Command, args []string) {
						verbose, _ := cmd.Flags().GetBool("verbose")
						if verbose {
							cmd.Print("Detailed listing: item1, item2, item3")
						} else {
							cmd.Print("item1 item2 item3")
						}
					},
				}
				cmd.Flags().Bool("verbose", false, "Show detailed output")
				return cmd
			},
			parentPath: "",
			args: map[string]interface{}{
				"verbose": true,
			},
			expectedOutput: "Detailed listing: item1, item2, item3",
			expectError:    false,
			description:    "Command with boolean flag should set the flag value correctly",
		},
		{
			name: "command with integer flag",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{
					Use:   "repeat",
					Short: "Repeat a message",
					Run: func(cmd *cobra.Command, args []string) {
						count, _ := cmd.Flags().GetInt("count")
						message := "hello"
						if len(args) > 0 {
							message = args[0]
						}
						for i := 0; i < count; i++ {
							if i > 0 {
								cmd.Print(" ")
							}
							cmd.Print(message)
						}
					},
				}
				cmd.Flags().Int("count", 1, "Number of times to repeat")
				return cmd
			},
			parentPath: "",
			args: map[string]interface{}{
				"count": float64(3), // JSON numbers are float64
				"args":  "test",
			},
			expectedOutput: "test test test",
			expectError:    false,
			description:    "Command with integer flag should convert and set the flag value correctly",
		},
		{
			name: "command with float flag",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{
					Use:   "calculate",
					Short: "Calculate something",
					Run: func(cmd *cobra.Command, args []string) {
						rate, _ := cmd.Flags().GetFloat64("rate")
						cmd.Printf("Rate: %.2f", rate)
					},
				}
				cmd.Flags().Float64("rate", 0.0, "The rate value")
				return cmd
			},
			parentPath: "",
			args: map[string]interface{}{
				"rate": 3.14,
			},
			expectedOutput: "Rate: 3.14",
			expectError:    false,
			description:    "Command with float flag should set the flag value correctly",
		},
		{
			name: "command that returns error",
			setupCmd: func() *cobra.Command {
				return &cobra.Command{
					Use:   "fail",
					Short: "A command that fails",
					RunE: func(cmd *cobra.Command, args []string) error {
						cmd.Print("This command failed")
						return errors.New("command failed")
					},
				}
			},
			parentPath:     "",
			args:           map[string]interface{}{},
			expectedOutput: "",
			expectError:    true,
			description:    "Command that returns an error should be handled properly",
		},
		{
			name: "command with no Run function",
			setupCmd: func() *cobra.Command {
				return &cobra.Command{
					Use:   "norun",
					Short: "A command with no run function",
					// No Run or RunE function
				}
			},
			parentPath:     "",
			args:           map[string]interface{}{},
			expectedOutput: "",
			expectError:    true,
			description:    "Command with no Run function should return an error",
		},
		{
			name: "subcommand execution",
			setupCmd: func() *cobra.Command {
				return &cobra.Command{
					Use:   "subcmd",
					Short: "A subcommand",
					Run: func(cmd *cobra.Command, args []string) {
						cmd.Print("Subcommand executed")
					},
				}
			},
			parentPath:     "parent",
			args:           map[string]interface{}{},
			expectedOutput: "Subcommand executed",
			expectError:    false,
			description:    "Subcommand should execute correctly with parent path",
		},
		{
			name: "command with persistent flags",
			setupCmd: func() *cobra.Command {
				// Create a parent command with persistent flags
				parent := &cobra.Command{
					Use:   "parent",
					Short: "Parent command",
				}
				parent.PersistentFlags().String("config", "default", "Config file")

				// Create child command
				child := &cobra.Command{
					Use:   "child",
					Short: "Child command",
					Run: func(cmd *cobra.Command, args []string) {
						config, _ := cmd.Parent().PersistentFlags().GetString("config")
						cmd.Printf("Using config: %s", config)
					},
				}

				parent.AddCommand(child)
				return child
			},
			parentPath: "parent",
			args: map[string]interface{}{
				"config": "custom.conf",
			},
			expectedOutput: "Using config: custom.conf",
			expectError:    false,
			description:    "Command with persistent flags from parent should work correctly",
		},
		{
			name: "command with hyphenated flag names",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{
					Use:   "deploy",
					Short: "Deploy application",
					Run: func(cmd *cobra.Command, args []string) {
						dryRun, _ := cmd.Flags().GetBool("dry-run")
						if dryRun {
							cmd.Print("Dry run mode")
						} else {
							cmd.Print("Deploying")
						}
					},
				}
				cmd.Flags().Bool("dry-run", false, "Perform a dry run")
				return cmd
			},
			parentPath: "",
			args: map[string]interface{}{
				"dry_run": true, // MCP converts hyphens to underscores
			},
			expectedOutput: "Dry run mode",
			expectError:    false,
			description:    "Command with hyphenated flag names should handle underscore conversion",
		},
		{
			name: "command with empty args parameter",
			setupCmd: func() *cobra.Command {
				return &cobra.Command{
					Use:   "echo",
					Short: "Echo arguments",
					Run: func(cmd *cobra.Command, args []string) {
						if len(args) == 0 {
							cmd.Print("No arguments")
						} else {
							cmd.Printf("Args: %s", strings.Join(args, " "))
						}
					},
				}
			},
			parentPath: "",
			args: map[string]interface{}{
				"args": "", // Empty string
			},
			expectedOutput: "No arguments",
			expectError:    false,
			description:    "Command with empty args parameter should handle gracefully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			bridge := NewCobraToMCPBridge(nil, "test", "1.0.0")
			cmd := tt.setupCmd()

			// Capture output
			var output strings.Builder
			cmd.SetOut(&output)
			cmd.SetErr(&output)

			// Create a mock CallToolRequest
			request := mcp.CallToolRequest{}
			request.Params.Arguments = tt.args

			// Execute
			result, err := bridge.executeCommand(context.Background(), cmd, tt.parentPath, request)

			// Verify error expectation
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
				return
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// If we expect an error, we're done
			if tt.expectError {
				return
			}

			// Verify result
			if result == nil {
				t.Errorf("Expected result but got nil")
				return
			}

			// For simplicity, let's check if the command actually produced the expected output
			// by examining what was written to the command's output
			actualOutput := output.String()

			if actualOutput != tt.expectedOutput {
				t.Errorf("Expected output '%s' but got '%s'", tt.expectedOutput, actualOutput)
			}
		})
	}
}

// TestCobraToMCPBridge_CreateMCPServer tests the CreateMCPServer method
func TestCobraToMCPBridge_CreateMCPServer(t *testing.T) {
	tests := []struct {
		name        string
		setupCmd    func() *cobra.Command
		appName     string
		version     string
		description string
	}{
		{
			name: "simple command registration",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{
					Use:   "test",
					Short: "Test command",
					Run: func(cmd *cobra.Command, args []string) {
						cmd.Print("test output")
					},
				}
				cmd.Flags().String("flag", "default", "A test flag")
				return cmd
			},
			appName:     "testapp",
			version:     "1.0.0",
			description: "Simple command should be registered as MCP tool",
		},
		{
			name: "command with subcommands",
			setupCmd: func() *cobra.Command {
				root := &cobra.Command{
					Use:   "root",
					Short: "Root command",
				}

				sub := &cobra.Command{
					Use:   "sub",
					Short: "Sub command",
					Run: func(cmd *cobra.Command, args []string) {
						cmd.Print("sub output")
					},
				}

				root.AddCommand(sub)
				return root
			},
			appName:     "testapp",
			version:     "1.0.0",
			description: "Command with subcommands should register subcommands as tools",
		},
		{
			name: "command with no run function but has subcommands",
			setupCmd: func() *cobra.Command {
				root := &cobra.Command{
					Use:   "root",
					Short: "Root command without run",
					// No Run function
				}

				sub1 := &cobra.Command{
					Use:   "sub1",
					Short: "Sub command 1",
					Run: func(cmd *cobra.Command, args []string) {
						cmd.Print("sub1 output")
					},
				}

				sub2 := &cobra.Command{
					Use:   "sub2",
					Short: "Sub command 2",
					Run: func(cmd *cobra.Command, args []string) {
						cmd.Print("sub2 output")
					},
				}

				root.AddCommand(sub1, sub2)
				return root
			},
			appName:     "testapp",
			version:     "1.0.0",
			description: "Root command without run function should only register subcommands",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			cmd := tt.setupCmd()
			bridge := NewCobraToMCPBridge(cmd, tt.appName, tt.version)

			// Execute
			server := bridge.CreateMCPServer()

			// Verify
			if server == nil {
				t.Errorf("Expected MCP server but got nil")
				return
			}

			// Verify the bridge has the server reference
			if bridge.server == nil {
				t.Errorf("Expected bridge to have server reference but got nil")
			}

			// Note: We can't easily test the internal tool registration without
			// exposing more internals, but we can at least verify the server was created
		})
	}
}

// TestCobraToMCPBridge_NewCobraToMCPBridge tests the constructor
func TestCobraToMCPBridge_NewCobraToMCPBridge(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	appName := "testapp"
	version := "1.0.0"

	bridge := NewCobraToMCPBridge(cmd, appName, version)

	if bridge == nil {
		t.Errorf("Expected bridge instance but got nil")
		return
	}

	if bridge.rootCmd != cmd {
		t.Errorf("Expected rootCmd to be set correctly")
	}

	if bridge.appName != appName {
		t.Errorf("Expected appName to be '%s' but got '%s'", appName, bridge.appName)
	}

	if bridge.version != version {
		t.Errorf("Expected version to be '%s' but got '%s'", version, bridge.version)
	}

	if bridge.server != nil {
		t.Errorf("Expected server to be nil initially but got %v", bridge.server)
	}
}

// TestCobraToMCPBridge_getCommandDescription tests the getCommandDescription method
func TestCobraToMCPBridge_getCommandDescription(t *testing.T) {
	tests := []struct {
		name         string
		setupCmd     func() *cobra.Command
		parentPath   string
		expectedDesc string
		description  string
	}{
		{
			name: "command with short description",
			setupCmd: func() *cobra.Command {
				return &cobra.Command{
					Use:   "test",
					Short: "Test command",
				}
			},
			parentPath:   "",
			expectedDesc: "Test command",
			description:  "Should use short description when available",
		},
		{
			name: "command with long description",
			setupCmd: func() *cobra.Command {
				return &cobra.Command{
					Use:   "test",
					Short: "Test command",
					Long:  "This is a longer description of the test command",
				}
			},
			parentPath:   "",
			expectedDesc: "Test command\n\nThis is a longer description of the test command",
			description:  "Should combine short and long descriptions",
		},
		{
			name: "command with Use information",
			setupCmd: func() *cobra.Command {
				return &cobra.Command{
					Use:   "test [args...]",
					Short: "Test command",
				}
			},
			parentPath:   "",
			expectedDesc: "Test command\n\nUsage: test [args...]",
			description:  "Should include usage information",
		},
		{
			name: "command with no descriptions",
			setupCmd: func() *cobra.Command {
				return &cobra.Command{
					Use: "test",
				}
			},
			parentPath:   "",
			expectedDesc: "Execute the 'test' command",
			description:  "Should generate default description when none provided",
		},
		{
			name: "subcommand with parent path",
			setupCmd: func() *cobra.Command {
				return &cobra.Command{
					Use:   "sub",
					Short: "Sub command",
				}
			},
			parentPath:   "parent_group",
			expectedDesc: "Sub command",
			description:  "Should use short description for subcommands",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bridge := NewCobraToMCPBridge(nil, "test", "1.0.0")
			cmd := tt.setupCmd()

			result := strings.TrimSpace(bridge.getCommandDescription(cmd, tt.parentPath))

			if result != tt.expectedDesc {
				t.Errorf("Expected description '%s' but got '%s'", tt.expectedDesc, result)
			}
		})
	}
}
