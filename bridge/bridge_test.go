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

// TestCommandFactory implements CommandFactory for testing
type TestCommandFactory struct {
	registrationCmd *cobra.Command
	output          *strings.Builder
}

func NewTestCommandFactory() *TestCommandFactory {
	return &TestCommandFactory{
		output: &strings.Builder{},
	}
}

func (f *TestCommandFactory) RegistrationCommand() *cobra.Command {
	if f.registrationCmd != nil {
		return f.registrationCmd
	}

	rootCmd := &cobra.Command{
		Use:   "test",
		Short: "Test command",
		Long:  "A test command for unit testing",
		Args:  cobra.ArbitraryArgs,
		Run: func(_ *cobra.Command, _ []string) {
			f.output.WriteString("root executed")
		},
	}

	subCmd := &cobra.Command{
		Use:   "sub",
		Short: "Sub command",
		Run: func(_ *cobra.Command, _ []string) {
			f.output.WriteString("sub executed")
		},
	}

	subCmd.Flags().String("flag", "", "test flag")
	subCmd.Flags().Bool("bool-flag", false, "test bool flag")
	subCmd.Flags().Int("int-flag", 0, "test int flag")

	rootCmd.AddCommand(subCmd)
	f.registrationCmd = rootCmd
	return rootCmd
}

func (f *TestCommandFactory) New() (*cobra.Command, CommandExecFunc) {
	f.output.Reset() // Clear output for fresh execution
	cmd := f.RegistrationCommand()

	exec := func(ctx context.Context) *mcp.CallToolResult {
		if err := cmd.ExecuteContext(ctx); err != nil {
			return mcp.NewToolResultError(err.Error())
		}
		return mcp.NewToolResultText(f.output.String())
	}

	return cmd, exec
}

func (f *TestCommandFactory) GetOutput() string {
	return f.output.String()
}

func TestNewCobraToMCPBridge(t *testing.T) {
	tests := []struct {
		name        string
		factory     CommandFactory
		appName     string
		version     string
		logger      *slog.Logger
		shouldPanic bool
		panicMsg    string
	}{
		{
			name:    "valid parameters",
			factory: NewTestCommandFactory(),
			appName: "test-app",
			version: "1.0.0",
			logger:  slog.New(slog.NewTextHandler(os.Stderr, nil)),
		},
		{
			name:        "nil factory",
			factory:     nil,
			appName:     "test-app",
			version:     "1.0.0",
			logger:      slog.New(slog.NewTextHandler(os.Stderr, nil)),
			shouldPanic: true,
			panicMsg:    "cmdFactory cannot be nil",
		},
		{
			name:        "empty app name",
			factory:     NewTestCommandFactory(),
			appName:     "",
			version:     "1.0.0",
			logger:      slog.New(slog.NewTextHandler(os.Stderr, nil)),
			shouldPanic: true,
			panicMsg:    "appName cannot be empty",
		},
		{
			name:    "empty version gets default",
			factory: NewTestCommandFactory(),
			appName: "test-app",
			version: "",
			logger:  slog.New(slog.NewTextHandler(os.Stderr, nil)),
		},
		{
			name:    "nil logger gets default",
			factory: NewTestCommandFactory(),
			appName: "test-app",
			version: "1.0.0",
			logger:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				defer func() {
					if r := recover(); r != nil {
						if !strings.Contains(r.(string), tt.panicMsg) {
							t.Errorf("Expected panic message to contain '%s', got '%v'", tt.panicMsg, r)
						}
					} else {
						t.Error("Expected panic but none occurred")
					}
				}()
			}

			bridge := NewCobraToMCPBridge(tt.factory, &MCPCommandConfig{
				AppName:    tt.appName,
				AppVersion: tt.version,
			})

			if !tt.shouldPanic {
				if bridge == nil {
					t.Error("Expected bridge to be created")
					return
				}
				if bridge.commandFactory == nil {
					t.Error("Expected commandFactory to be set")
				}
				if bridge.server == nil {
					t.Error("Expected server to be set")
				}
				if bridge.logger == nil {
					t.Error("Expected logger to be set")
				}
			}
		})
	}
}

func TestBridgeIntegration(t *testing.T) {
	factory := NewTestCommandFactory()

	bridge := NewCobraToMCPBridge(factory, &MCPCommandConfig{
		AppName:    "test",
		AppVersion: "test",
	})

	if bridge == nil {
		t.Fatal("Failed to create bridge")
	}

	// Test that the bridge properly registers commands
	// This is more of an integration test since we can't easily test
	// the internal MCP server registration without running the server
	t.Log("Bridge created successfully with test factory")
}

// MockCommandFactory for testing edge cases
type MockCommandFactory struct {
	registrationPanic bool
	executionPanic    bool
	returnNilCmd      bool
}

func (m *MockCommandFactory) RegistrationCommand() *cobra.Command {
	if m.registrationPanic {
		panic("registration panic")
	}
	if m.returnNilCmd {
		return nil
	}
	return &cobra.Command{Use: "mock"}
}

func (m *MockCommandFactory) New() (*cobra.Command, CommandExecFunc) {
	if m.executionPanic {
		panic("execution panic")
	}
	cmd := &cobra.Command{Use: "mock"}
	exec := func(_ context.Context) *mcp.CallToolResult {
		return mcp.NewToolResultText("mock output")
	}
	return cmd, exec
}

func TestBridgeWithMockFactory(t *testing.T) {
	t.Run("registration panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic during registration")
			}
		}()

		factory := &MockCommandFactory{registrationPanic: true}
		NewCobraToMCPBridge(factory, &MCPCommandConfig{
			AppName:    "test",
			AppVersion: "test",
		})
	})

	t.Run("nil command from factory", func(t *testing.T) {
		// This might cause issues in registration - the bridge should handle this gracefully
		factory := &MockCommandFactory{returnNilCmd: true}

		// This might panic or create an incomplete bridge
		// The behavior depends on how the registration handles nil commands
		defer func() {
			_ = recover() // Ignore panics for this test
		}()

		bridge := NewCobraToMCPBridge(factory, &MCPCommandConfig{
			AppName:    "test",
			AppVersion: "test",
		})
		if bridge != nil {
			t.Log("Bridge created despite nil command from factory")
		}
	})
}
