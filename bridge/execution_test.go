package bridge

import (
	"context"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/njayp/ophis/tools"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecuteCommand(t *testing.T) {
	tests := []struct {
		name           string
		toolName       string
		arguments      map[string]any
		expectedOutput string
		expectedError  bool
	}{
		{
			name:     "simple command execution",
			toolName: "test",
			arguments: map[string]any{
				tools.FlagsParam:          map[string]any{},
				tools.PositionalArgsParam: "",
			},
			expectedOutput: "test executed",
			expectedError:  false,
		},
		{
			name:     "command with flags",
			toolName: "test",
			arguments: map[string]any{
				tools.FlagsParam: map[string]any{
					"verbose": true,
				},
				tools.PositionalArgsParam: "",
			},
			expectedOutput: "test executed with verbose",
			expectedError:  false,
		},
		{
			name:     "command with positional args",
			toolName: "test",
			arguments: map[string]any{
				tools.FlagsParam:          map[string]any{},
				tools.PositionalArgsParam: "arg1 arg2",
			},
			expectedOutput: "test executed with args: arg1 arg2",
			expectedError:  false,
		},
		{
			name:     "command with quoted positional args",
			toolName: "test",
			arguments: map[string]any{
				tools.FlagsParam:          map[string]any{},
				tools.PositionalArgsParam: `"hello world" 'single quoted' unquoted`,
			},
			expectedOutput: "test executed with args: hello world single quoted unquoted",
			expectedError:  false,
		},
		{
			name:     "command with malformed quotes",
			toolName: "test",
			arguments: map[string]any{
				tools.FlagsParam:          map[string]any{},
				tools.PositionalArgsParam: `arg1 "unterminated quote`,
			},
			expectedOutput: "test executed with args: arg1 \"unterminated quote",
			expectedError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test command factory
			factory := &testCommandFactory{}

			// Create manager
			config := &Config{
				AppName:    "test-app",
				AppVersion: "1.0.0",
			}

			manager, err := New(factory, config)
			require.NoError(t, err)

			// Create test tool
			tool := tools.Tool{
				Tool: mcp.NewTool(tt.toolName),
			}

			// Create mock request
			request := mcp.CallToolRequest{}
			request.Params.Arguments = tt.arguments

			// Execute command
			result := manager.executeCommand(context.Background(), tool, request)

			if tt.expectedError {
				assert.True(t, result.IsError)
			} else {
				assert.False(t, result.IsError)
				if len(result.Content) > 0 {
					if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
						assert.Contains(t, textContent.Text, tt.expectedOutput)
					}
				}
			}
		})
	}
}

// testCommandFactory implements CommandFactory for testing
type testCommandFactory struct{}

func (f *testCommandFactory) Tools() []tools.Tool {
	cmd := f.createTestCommand()
	return tools.FromRootCmd(cmd)
}

func (f *testCommandFactory) New() (*cobra.Command, CommandExecFunc) {
	rcmd := f.createTestCommand()

	execFunc := func(ctx context.Context, cmd *cobra.Command) *mcp.CallToolResult {
		var output strings.Builder
		cmd.SetOut(&output)
		cmd.SetErr(&output)

		err := cmd.ExecuteContext(ctx)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("Command failed", err)
		}

		return mcp.NewToolResultText(output.String())
	}

	return rcmd, execFunc
}

func (f *testCommandFactory) createTestCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "test",
		Run: func(cmd *cobra.Command, args []string) {
			output := "test executed"

			if verbose, _ := cmd.Flags().GetBool("verbose"); verbose {
				output += " with verbose"
			}

			if len(args) > 0 {
				output += " with args: " + strings.Join(args, " ")
			}

			cmd.Print(output)
		},
	}

	cmd.Flags().Bool("verbose", false, "Enable verbose output")
	return cmd
}

func TestDescendCmdTree(t *testing.T) {
	tests := []struct {
		name        string
		cmdPath     []string
		expectError bool
	}{
		{
			name:        "root command",
			cmdPath:     []string{"test"},
			expectError: false,
		},
		{
			name:        "subcommand",
			cmdPath:     []string{"test", "sub"},
			expectError: false,
		},
		{
			name:        "invalid path",
			cmdPath:     []string{"test", "nonexistent"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test command with subcommand
			rootCmd := &cobra.Command{Use: "test"}
			subCmd := &cobra.Command{Use: "sub"}
			rootCmd.AddCommand(subCmd)

			result, err := descendCmdTree(rootCmd, tt.cmdPath)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}
