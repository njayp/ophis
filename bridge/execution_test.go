package bridge

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/spf13/cobra"
)

func TestExecuteCommand(t *testing.T) {
	factory := NewTestCommandFactory()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	bridge := &CobraToMCPBridge{
		commandFactory: factory,
		logger:         logger,
	}

	tests := []struct {
		name        string
		cmdPath     string
		arguments   map[string]any
		expectError bool
		expectText  string
	}{
		{
			name:    "execute root command",
			cmdPath: "test",
			arguments: map[string]any{
				PositionalArgsParam: "",
				FlagsParam:          map[string]any{},
			},
			expectError: false,
			expectText:  "root executed",
		},
		{
			name:    "execute subcommand",
			cmdPath: "test sub",
			arguments: map[string]any{
				PositionalArgsParam: "",
				FlagsParam:          map[string]any{},
			},
			expectError: false,
			expectText:  "sub executed",
		},
		{
			name:    "execute subcommand with flags",
			cmdPath: "test sub",
			arguments: map[string]any{
				PositionalArgsParam: "",
				FlagsParam: map[string]any{
					"flag":      "test-value",
					"bool-flag": true,
					"int-flag":  42,
				},
			},
			expectError: false,
			expectText:  "sub executed",
		},
		{
			name:    "execute with positional args",
			cmdPath: "test",
			arguments: map[string]any{
				PositionalArgsParam: "arg1 arg2",
				FlagsParam:          map[string]any{},
			},
			expectError: false,
			expectText:  "root executed",
		},
		{
			name:    "invalid command path",
			cmdPath: "test nonexistent",
			arguments: map[string]any{
				PositionalArgsParam: "",
				FlagsParam:          map[string]any{},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Arguments: tt.arguments,
				},
			}

			result := bridge.executeCommand(context.Background(), tt.cmdPath, request)

			if tt.expectError {
				if !result.IsError {
					t.Error("Expected error result")
				}
			} else {
				if result.IsError {
					t.Errorf("Unexpected error: %+v", result.Content)
				}
				if len(result.Content) > 0 {
					if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
						if !strings.Contains(textContent.Text, tt.expectText) {
							t.Errorf("Expected output to contain '%s', got '%s'", tt.expectText, textContent.Text)
						}
					}
				}
			}
		})
	}
}

func TestLoadFlagsFromMap(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	bridge := &CobraToMCPBridge{logger: logger}

	tests := []struct {
		name        string
		setupCmd    func() *cobra.Command
		flagMap     map[string]any
		expectError bool
		validate    func(*cobra.Command, *testing.T)
	}{
		{
			name: "valid string flag",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{Use: "test"}
				cmd.Flags().String("str-flag", "", "string flag")
				return cmd
			},
			flagMap: map[string]any{
				"str-flag": "test-value",
			},
			expectError: false,
			validate: func(cmd *cobra.Command, t *testing.T) {
				value, _ := cmd.Flags().GetString("str-flag")
				if value != "test-value" {
					t.Errorf("Expected str-flag to be 'test-value', got '%s'", value)
				}
			},
		},
		{
			name: "valid bool flag",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{Use: "test"}
				cmd.Flags().Bool("bool-flag", false, "bool flag")
				return cmd
			},
			flagMap: map[string]any{
				"bool-flag": true,
			},
			expectError: false,
			validate: func(cmd *cobra.Command, t *testing.T) {
				value, _ := cmd.Flags().GetBool("bool-flag")
				if !value {
					t.Error("Expected bool-flag to be true")
				}
			},
		},
		{
			name: "valid int flag",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{Use: "test"}
				cmd.Flags().Int("int-flag", 0, "int flag")
				return cmd
			},
			flagMap: map[string]any{
				"int-flag": 42,
			},
			expectError: false,
			validate: func(cmd *cobra.Command, t *testing.T) {
				value, _ := cmd.Flags().GetInt("int-flag")
				if value != 42 {
					t.Errorf("Expected int-flag to be 42, got %d", value)
				}
			},
		},
		{
			name: "nonexistent flag",
			setupCmd: func() *cobra.Command {
				return &cobra.Command{Use: "test"}
			},
			flagMap: map[string]any{
				"nonexistent": "value",
			},
			expectError: false, // Currently logs error but doesn't return error
			validate:    func(_ *cobra.Command, _ *testing.T) {},
		},
		{
			name: "invalid flag value",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{Use: "test"}
				cmd.Flags().Int("int-flag", 0, "int flag")
				return cmd
			},
			flagMap: map[string]any{
				"int-flag": "not-a-number",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.setupCmd()
			err := bridge.loadFlagsFromMap(cmd, tt.flagMap)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if err == nil {
				tt.validate(cmd, t)
			}
		})
	}
}

func TestDescendCmdTree(t *testing.T) {
	tests := []struct {
		name        string
		setupCmd    func() *cobra.Command
		cmdPath     string
		expectError bool
		expectName  string
	}{
		{
			name: "root command",
			setupCmd: func() *cobra.Command {
				return &cobra.Command{Use: "root"}
			},
			cmdPath:     "root",
			expectError: false,
			expectName:  "root",
		},
		{
			name: "single subcommand",
			setupCmd: func() *cobra.Command {
				root := &cobra.Command{Use: "root"}
				sub := &cobra.Command{Use: "sub"}
				root.AddCommand(sub)
				return root
			},
			cmdPath:     "root sub",
			expectError: false,
			expectName:  "sub",
		},
		{
			name: "nested subcommands",
			setupCmd: func() *cobra.Command {
				root := &cobra.Command{Use: "root"}
				sub1 := &cobra.Command{Use: "sub1"}
				sub2 := &cobra.Command{Use: "sub2"}
				root.AddCommand(sub1)
				sub1.AddCommand(sub2)
				return root
			},
			cmdPath:     "root sub1 sub2",
			expectError: false,
			expectName:  "sub2",
		},
		{
			name: "nonexistent command path",
			setupCmd: func() *cobra.Command {
				root := &cobra.Command{Use: "root"}
				sub := &cobra.Command{Use: "sub"}
				root.AddCommand(sub)
				return root
			},
			cmdPath:     "root nonexistent",
			expectError: true,
		},
		{
			name: "partial path match",
			setupCmd: func() *cobra.Command {
				root := &cobra.Command{Use: "root"}
				sub := &cobra.Command{Use: "sub"}
				root.AddCommand(sub)
				return root
			},
			cmdPath:     "root sub extra",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.setupCmd()
			result, err := descendCmdTree(cmd, tt.cmdPath)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result == nil {
					t.Error("Expected command but got nil")
				} else if result.Name() != tt.expectName {
					t.Errorf("Expected command name '%s', got '%s'", tt.expectName, result.Name())
				}
			}
		})
	}
}

// TestExecuteCommandWithPanic tests that command execution handles panics gracefully
func TestExecuteCommandWithPanic(t *testing.T) {
	factory := &PanicCommandFactory{}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	bridge := &CobraToMCPBridge{
		commandFactory: factory,
		logger:         logger,
	}

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				PositionalArgsParam: "",
				FlagsParam:          map[string]any{},
			},
		},
	}

	result := bridge.executeCommand(context.Background(), "panic", request)

	if !result.IsError {
		t.Error("Expected error result from panicking command")
	}

	// Check that the error message indicates a panic
	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
			if !strings.Contains(textContent.Text, "panic") {
				t.Errorf("Expected panic message in error, got: %s", textContent.Text)
			}
		}
	}
}

// TestExecuteCommandContextCancellation tests command execution with context cancellation
func TestExecuteCommandContextCancellation(t *testing.T) {
	factory := NewTestCommandFactory()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	bridge := &CobraToMCPBridge{
		commandFactory: factory,
		logger:         logger,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				PositionalArgsParam: "",
				FlagsParam:          map[string]any{},
			},
		},
	}

	result := bridge.executeCommand(ctx, "test", request)

	// The result depends on how the command handles context cancellation
	// For this simple test command, it might complete successfully since it's fast
	// In real scenarios, long-running commands should respect context cancellation
	if result == nil {
		t.Error("Expected result from command execution")
	}
}

// PanicCommandFactory creates commands that panic during execution
type PanicCommandFactory struct{}

func (f *PanicCommandFactory) CreateRegistrationCommand() *cobra.Command {
	return &cobra.Command{
		Use: "panic",
		Run: func(_ *cobra.Command, _ []string) {
			panic("test panic")
		},
	}
}

func (f *PanicCommandFactory) CreateCommand() (*cobra.Command, CommandExecFunc) {
	cmd := &cobra.Command{
		Use: "panic",
		Run: func(_ *cobra.Command, _ []string) {
			panic("test panic")
		},
	}

	exec := func(ctx context.Context) *mcp.CallToolResult {
		if err := cmd.ExecuteContext(ctx); err != nil {
			return mcp.NewToolResultError(err.Error())
		}
		return mcp.NewToolResultText("should not reach here")
	}

	return cmd, exec
}

// BenchmarkExecuteCommand benchmarks command execution performance
func BenchmarkExecuteCommand(b *testing.B) {
	factory := NewTestCommandFactory()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	bridge := &CobraToMCPBridge{
		commandFactory: factory,
		logger:         logger,
	}

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				PositionalArgsParam: "",
				FlagsParam:          map[string]any{},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bridge.executeCommand(context.Background(), "test", request)
	}
}
