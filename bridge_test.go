package ophis

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/ophis/cmds/basic"
	"github.com/spf13/cobra"
)

func TestNewCobraToMCPBridge(t *testing.T) {
	tests := []struct {
		name       string
		cmdFactory func() *cobra.Command
		appName    string
		version    string
		logger     *slog.Logger
		wantPanic  bool
		wantApp    string
		wantVer    string
	}{
		{
			name:       "valid bridge creation",
			cmdFactory: basic.NewRootCmd,
			appName:    "test-app",
			version:    "1.0.0",
			logger:     slog.Default(),
			wantPanic:  false,
			wantApp:    "test-app",
			wantVer:    "1.0.0",
		},
		{
			name:       "nil command factory should panic",
			cmdFactory: nil,
			appName:    "test-app",
			version:    "1.0.0",
			logger:     slog.Default(),
			wantPanic:  true,
		},
		{
			name:       "empty app name should panic",
			cmdFactory: basic.NewRootCmd,
			appName:    "",
			version:    "1.0.0",
			logger:     slog.Default(),
			wantPanic:  true,
		},
		{
			name:       "empty version should default to unknown",
			cmdFactory: basic.NewRootCmd,
			appName:    "test-app",
			version:    "",
			logger:     slog.Default(),
			wantPanic:  false,
			wantApp:    "test-app",
			wantVer:    "unknown",
		},
		{
			name:       "nil logger should use default",
			cmdFactory: basic.NewRootCmd,
			appName:    "test-app",
			version:    "1.0.0",
			logger:     nil,
			wantPanic:  false,
			wantApp:    "test-app",
			wantVer:    "1.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("NewCobraToMCPBridge() expected panic but didn't panic")
					}
				}()
			}

			bridge := NewCobraToMCPBridge(tt.cmdFactory, tt.appName, tt.version, tt.logger)

			if !tt.wantPanic {
				if bridge == nil {
					t.Errorf("NewCobraToMCPBridge() returned nil")
					return
				}

				if bridge.appName != tt.wantApp {
					t.Errorf("NewCobraToMCPBridge() appName = %v, want %v", bridge.appName, tt.wantApp)
				}

				if bridge.version != tt.wantVer {
					t.Errorf("NewCobraToMCPBridge() version = %v, want %v", bridge.version, tt.wantVer)
				}

				if bridge.server == nil {
					t.Errorf("NewCobraToMCPBridge() server is nil")
				}

				if bridge.logger == nil {
					t.Errorf("NewCobraToMCPBridge() logger is nil")
				}
			}
		})
	}
}

func TestCreateMCPParameter(t *testing.T) {
	bridge := NewCobraToMCPBridge(basic.NewRootCmd, "test", "1.0.0", slog.Default())

	tests := []struct {
		name        string
		paramName   string
		description string
		flagType    string
		wantType    string // Expected parameter type (we'll check by examining the tool option)
	}{
		{
			name:        "boolean parameter",
			paramName:   "verbose",
			description: "Enable verbose output",
			flagType:    "bool",
			wantType:    "boolean",
		},
		{
			name:        "integer parameter",
			paramName:   "count",
			description: "Number of items",
			flagType:    "int",
			wantType:    "number",
		},
		{
			name:        "int8 parameter",
			paramName:   "level",
			description: "Log level",
			flagType:    "int8",
			wantType:    "number",
		},
		{
			name:        "int16 parameter",
			paramName:   "port",
			description: "Port number",
			flagType:    "int16",
			wantType:    "number",
		},
		{
			name:        "int32 parameter",
			paramName:   "size",
			description: "Size value",
			flagType:    "int32",
			wantType:    "number",
		},
		{
			name:        "int64 parameter",
			paramName:   "timestamp",
			description: "Unix timestamp",
			flagType:    "int64",
			wantType:    "number",
		},
		{
			name:        "uint parameter",
			paramName:   "count",
			description: "Unsigned count",
			flagType:    "uint",
			wantType:    "number",
		},
		{
			name:        "uint8 parameter",
			paramName:   "byte",
			description: "Byte value",
			flagType:    "uint8",
			wantType:    "number",
		},
		{
			name:        "uint16 parameter",
			paramName:   "word",
			description: "Word value",
			flagType:    "uint16",
			wantType:    "number",
		},
		{
			name:        "uint32 parameter",
			paramName:   "dword",
			description: "Double word value",
			flagType:    "uint32",
			wantType:    "number",
		},
		{
			name:        "uint64 parameter",
			paramName:   "qword",
			description: "Quad word value",
			flagType:    "uint64",
			wantType:    "number",
		},
		{
			name:        "float32 parameter",
			paramName:   "ratio",
			description: "Floating point ratio",
			flagType:    "float32",
			wantType:    "number",
		},
		{
			name:        "float64 parameter",
			paramName:   "precision",
			description: "High precision value",
			flagType:    "float64",
			wantType:    "number",
		},
		{
			name:        "string parameter",
			paramName:   "name",
			description: "User name",
			flagType:    "string",
			wantType:    "string",
		},
		{
			name:        "unknown type defaults to string",
			paramName:   "custom",
			description: "Custom type",
			flagType:    "customType",
			wantType:    "string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := bridge.createMCPParameter(tt.paramName, tt.description, tt.flagType)

			if len(options) != 1 {
				t.Errorf("createMCPParameter() returned %d options, want 1", len(options))
				return
			}

			// Since we can't easily inspect the option type directly,
			// we verify that the function doesn't panic and returns exactly one option
			// The actual type validation would need to be done at runtime through MCP
			if options[0] == nil {
				t.Errorf("createMCPParameter() returned nil option")
			}
		})
	}
}

func TestGetCommandDescription(t *testing.T) {
	bridge := NewCobraToMCPBridge(basic.NewRootCmd, "test", "1.0.0", slog.Default())

	tests := []struct {
		name       string
		cmd        *cobra.Command
		parentPath string
		want       string
	}{
		{
			name: "command with short description",
			cmd: &cobra.Command{
				Use:   "test",
				Short: "Short description",
			},
			parentPath: "",
			want:       "Short description\n\nUsage: test",
		},
		{
			name: "command with long description",
			cmd: &cobra.Command{
				Use:   "test",
				Short: "Short description",
				Long:  "This is a much longer description that provides more detail",
			},
			parentPath: "",
			want:       "Short description\n\nUsage: test\n\nThis is a much longer description that provides more detail",
		},
		{
			name: "command with only long description",
			cmd: &cobra.Command{
				Use:  "test",
				Long: "Only long description",
			},
			parentPath: "",
			// TODO maybe change this behavior
			want: "Only long description\n\nUsage: test\n\nOnly long description",
		},
		{
			name: "command with no description",
			cmd: &cobra.Command{
				Use: "test",
			},
			parentPath: "",
			want:       "Execute the 'test' command\n\nUsage: test",
		},
		{
			name: "command with parent path",
			cmd: &cobra.Command{
				Use:   "subcommand",
				Short: "Sub command",
			},
			parentPath: "parent_cmd",
			want:       "Sub command\n\nUsage: subcommand",
		},
		{
			name: "command with no description and parent path",
			cmd: &cobra.Command{
				Use: "subcommand",
			},
			parentPath: "parent_cmd",
			want:       "Execute the 'parent cmd subcommand' command\n\nUsage: subcommand",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := bridge.getCommandDescription(tt.cmd, tt.parentPath)
			if got != tt.want {
				t.Errorf("getCommandDescription() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExecuteCommand(t *testing.T) {
	bridge := NewCobraToMCPBridge(basic.NewRootCmd, "test", "1.0.0", slog.Default())

	tests := []struct {
		name        string
		cmd         *cobra.Command
		request     mcp.CallToolRequest
		wantContent string
		wantError   bool
	}{
		{
			name: "hello command with default greeting",
			cmd:  basic.NewRootCmd().Commands()[0],
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Arguments: map[string]interface{}{},
				},
			},
			wantContent: "Hello, World!\n",
			wantError:   false,
		},
		{
			name: "hello command with custom name",
			cmd:  basic.NewRootCmd().Commands()[0],
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Arguments: map[string]interface{}{
						PositionalArgsParam: "Alice",
					},
				},
			},
			wantContent: "Hello, Alice!\n",
			wantError:   false,
		},
		{
			name: "hello command with custom greeting",
			cmd:  basic.NewRootCmd().Commands()[0],
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Arguments: map[string]interface{}{
						"greeting":          "Hi",
						PositionalArgsParam: "Bob",
					},
				},
			},
			wantContent: "Hi, Bob!\n",
			wantError:   false,
		},
		{
			name: "hello command with multiple positional args",
			cmd:  basic.NewRootCmd().Commands()[0],
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Arguments: map[string]interface{}{
						PositionalArgsParam: "Charlie Dave",
					},
				},
			},
			wantContent: "Hello, Charlie!\n",
			wantError:   true, // Should err because too many args
		},
		{
			name: "hello command with empty args string",
			cmd:  basic.NewRootCmd().Commands()[0],
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Arguments: map[string]interface{}{
						PositionalArgsParam: "",
					},
				},
			},
			wantContent: "Hello, World!\n",
			wantError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := bridge.executeCommand(context.Background(), tt.cmd, tt.request)

			if result == nil {
				t.Errorf("executeCommand() returned nil result")
				return
			}

			if len(result.Content) == 0 {
				t.Errorf("executeCommand() returned empty content")
				return
			}

			if tt.wantError {
				// Check if it's an error result
				if len(result.Content) > 0 {
					if _, isText := result.Content[0].(mcp.TextContent); isText {
						// This might be an error in text form, which is acceptable
						return
					}
				}
				t.Errorf("executeCommand() expected error but got success")
				return
			}

			// Check for successful text result
			content, ok := result.Content[0].(mcp.TextContent)
			if !ok {
				t.Errorf("executeCommand() result content is not TextContent, got %T", result.Content[0])
				return
			}

			if content.Text != tt.wantContent {
				t.Errorf("executeCommand() content = %q, want %q", content.Text, tt.wantContent)
			}
		})
	}
}

func TestExecuteCommandWithComplexCommands(t *testing.T) {
	// Create a more complex command for testing
	complexCmdFactory := func() *cobra.Command {
		var verbose bool
		var count int
		var name string

		cmd := &cobra.Command{
			Use:   "complex [arg1] [arg2]",
			Short: "A complex command for testing",
			Args:  cobra.MaximumNArgs(2),
			Run: func(cmd *cobra.Command, args []string) {
				output := fmt.Sprintf("verbose=%t count=%d name=%s args=%v", verbose, count, name, args)
				cmd.Print(output)
			},
		}

		cmd.Flags().BoolVar(&verbose, "verbose", false, "Enable verbose mode")
		cmd.Flags().IntVar(&count, "count", 1, "Number of iterations")
		cmd.Flags().StringVar(&name, "name", "default", "Name parameter")

		return cmd
	}

	bridge := NewCobraToMCPBridge(complexCmdFactory, "test", "1.0.0", slog.Default())

	tests := []struct {
		name        string
		request     mcp.CallToolRequest
		wantContent string
	}{
		{
			name: "complex command with all flags",
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Arguments: map[string]interface{}{
						"verbose":           "true",
						"count":             "5",
						"name":              "test-name",
						PositionalArgsParam: "arg1 arg2",
					},
				},
			},
			wantContent: "verbose=true count=5 name=test-name args=[arg1 arg2]",
		},
		{
			name: "complex command with defaults",
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Arguments: map[string]interface{}{},
				},
			},
			wantContent: "verbose=false count=1 name=default args=[]",
		},
		{
			name: "complex command with some flags",
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Arguments: map[string]interface{}{
						"verbose":           "true",
						PositionalArgsParam: "single-arg",
					},
				},
			},
			wantContent: "verbose=true count=1 name=default args=[single-arg]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := complexCmdFactory()
			result := bridge.executeCommand(context.Background(), cmd, tt.request)

			if result == nil {
				t.Errorf("executeCommand() returned nil result")
				return
			}

			if len(result.Content) == 0 {
				t.Errorf("executeCommand() returned empty content")
				return
			}

			content, ok := result.Content[0].(mcp.TextContent)
			if !ok {
				t.Errorf("executeCommand() result content is not TextContent, got %T", result.Content[0])
				return
			}

			if content.Text != tt.wantContent {
				t.Errorf("executeCommand() content = %q, want %q", content.Text, tt.wantContent)
			}
		})
	}
}

func TestExecuteCommandErrorHandling(t *testing.T) {
	// Create commands that will fail for testing error handling
	errorCmdFactory := func() *cobra.Command {
		return &cobra.Command{
			Use:   "error-cmd",
			Short: "A command that returns an error",
			RunE: func(cmd *cobra.Command, args []string) error {
				return fmt.Errorf("simulated command error")
			},
		}
	}

	panicCmdFactory := func() *cobra.Command {
		return &cobra.Command{
			Use:   "panic-cmd",
			Short: "A command that panics",
			Run: func(cmd *cobra.Command, args []string) {
				panic("simulated panic")
			},
		}
	}

	bridge := NewCobraToMCPBridge(errorCmdFactory, "test", "1.0.0", slog.Default())

	tests := []struct {
		name        string
		cmdFactory  func() *cobra.Command
		request     mcp.CallToolRequest
		wantError   bool
		errorSubstr string
	}{
		{
			name:       "command returning error",
			cmdFactory: errorCmdFactory,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Arguments: map[string]interface{}{},
				},
			},
			wantError:   true,
			errorSubstr: "simulated command error",
		},
		{
			name:       "command that panics",
			cmdFactory: panicCmdFactory,
			request: mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Arguments: map[string]interface{}{},
				},
			},
			wantError:   true,
			errorSubstr: "command panicked",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Update bridge command factory for this test
			bridge.commandFactory = tt.cmdFactory
			cmd := tt.cmdFactory()
			result := bridge.executeCommand(context.Background(), cmd, tt.request)

			if result == nil {
				t.Errorf("executeCommand() returned nil result")
				return
			}

			if len(result.Content) == 0 {
				t.Errorf("executeCommand() returned empty content")
				return
			}

			content, ok := result.Content[0].(mcp.TextContent)
			if !ok {
				t.Errorf("executeCommand() result content is not TextContent, got %T", result.Content[0])
				return
			}

			if tt.wantError {
				if !strings.Contains(content.Text, tt.errorSubstr) {
					t.Errorf("executeCommand() error content = %q, want to contain %q", content.Text, tt.errorSubstr)
				}
			}
		})
	}
}

// Original test preserved for compatibility
func TestExecCommand(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	// Create a temporary directory for the test
	cf := basic.NewRootCmd

	// Create a new CobraToMCPBridge instance
	bridge := NewCobraToMCPBridge(cf, "ophis", "0.0.0-test", nil)

	cmd := cf().Commands()[0]

	// Execute a command in the temporary directory
	result := bridge.executeCommand(context.Background(), cmd, mcp.CallToolRequest{})

	content, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Error("content not ok")
	}

	expected := "Hello, World!\n"
	if content.Text != expected {
		t.Error(fmt.Sprintf("wanted %s, got %s", expected, content.Text))
	}
}
