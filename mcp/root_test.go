package mcp

import (
	"testing"

	"github.com/njayp/ophis/bridge"
	"github.com/njayp/ophis/tools"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockCommandFactory implements CommandFactory for testing
type mockCommandFactory struct {
	tools []tools.Tool
	cmd   *cobra.Command
	exec  bridge.CommandExecFunc
}

func (m *mockCommandFactory) Tools() []tools.Tool {
	return m.tools
}

func (m *mockCommandFactory) New() (*cobra.Command, bridge.CommandExecFunc) {
	return m.cmd, m.exec
}

func TestCommand(t *testing.T) {
	tests := []struct {
		name              string
		factory           bridge.CommandFactory
		config            *bridge.Config
		expectedCommands  []string
		expectedUse       string
	}{
		{
			name: "creates mcp command with all subcommands",
			factory: &mockCommandFactory{
				tools: []tools.Tool{},
				cmd:   &cobra.Command{Use: "test"},
				exec:  nil,
			},
			config: &bridge.Config{
				AppName:    "test-app",
				AppVersion: "1.0.0",
			},
			expectedCommands: []string{"start", "tools", "claude"},
			expectedUse:      "mcp",
		},
		{
			name: "creates mcp command with custom config",
			factory: &mockCommandFactory{
				tools: []tools.Tool{},
				cmd:   &cobra.Command{Use: "custom"},
				exec:  nil,
			},
			config: &bridge.Config{
				AppName:    "custom-app",
				AppVersion: "2.0.0",
				LogLevel:   "debug",
				LogFile:    "/tmp/custom.log",
			},
			expectedCommands: []string{"start", "tools", "claude"},
			expectedUse:      "mcp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := Command(tt.factory, tt.config)

			// Verify command structure
			assert.Equal(t, tt.expectedUse, cmd.Use)
			assert.Equal(t, tools.MCPCommandName, cmd.Use)

			// Verify subcommands
			subcommands := cmd.Commands()
			assert.Len(t, subcommands, len(tt.expectedCommands))

			// Check each expected subcommand exists
			commandMap := make(map[string]*cobra.Command)
			for _, subcmd := range subcommands {
				commandMap[subcmd.Use] = subcmd
			}

			for _, expectedCmd := range tt.expectedCommands {
				_, exists := commandMap[expectedCmd]
				assert.True(t, exists, "Expected subcommand %s not found", expectedCmd)
			}

			// Verify the claude subcommand has its own subcommands
			claudeCmd, exists := commandMap["claude"]
			require.True(t, exists)
			claudeSubcommands := claudeCmd.Commands()
			assert.GreaterOrEqual(t, len(claudeSubcommands), 3) // enable, disable, list
		})
	}
}

func TestCommandIntegration(t *testing.T) {
	// Test that the command can be added to a root command and executed
	rootCmd := &cobra.Command{
		Use:   "test-cli",
		Short: "Test CLI",
	}

	factory := &mockCommandFactory{
		tools: []tools.Tool{},
		cmd:   &cobra.Command{Use: "mock"},
		exec:  nil,
	}

	config := &bridge.Config{
		AppName:    "test-cli",
		AppVersion: "1.0.0",
	}

	mcpCmd := Command(factory, config)
	rootCmd.AddCommand(mcpCmd)

	// Verify the command was added properly
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "mcp" {
			found = true
			break
		}
	}
	assert.True(t, found, "MCP command not found in root command")

	// Test command path resolution
	cmd, _, err := rootCmd.Find([]string{"mcp", "tools"})
	assert.NoError(t, err)
	assert.NotNil(t, cmd)
	assert.Equal(t, "tools", cmd.Use)
}

func TestCommandFactoryNil(t *testing.T) {
	// Test that nil factory is handled gracefully
	defer func() {
		if r := recover(); r != nil {
			// Expected to panic or handle nil gracefully
			t.Logf("Recovered from panic as expected: %v", r)
		}
	}()

	config := &bridge.Config{
		AppName:    "test",
		AppVersion: "1.0.0",
	}

	// This should be handled gracefully by the bridge.New() function
	// when the start command is executed
	cmd := Command(nil, config)
	assert.NotNil(t, cmd)
}
