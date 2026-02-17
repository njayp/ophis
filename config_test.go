package ophis

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestCmdFilter(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
		cmd      *cobra.Command
	}{
		{
			name:     "passing cmd",
			expected: false,
			cmd: &cobra.Command{
				Use: "test",
				Run: func(_ *cobra.Command, _ []string) {},
			},
		},
		{
			name:     "depreciated cmd",
			expected: true,
			cmd: &cobra.Command{
				Use:        "test",
				Run:        func(_ *cobra.Command, _ []string) {},
				Deprecated: "test",
			},
		},
		{
			name:     "hidden cmd",
			expected: true,
			cmd: &cobra.Command{
				Use:    "test",
				Run:    func(_ *cobra.Command, _ []string) {},
				Hidden: true,
			},
		},
		{
			name:     "mcp cmd",
			expected: true,
			cmd: &cobra.Command{
				Use: "mcp",
				Run: func(_ *cobra.Command, _ []string) {},
			},
		},
		{
			name:     "no run cmd",
			expected: true,
			cmd: &cobra.Command{
				Use: "test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Default config uses "mcp" as command name.
			c := &Config{}
			result := c.cmdFilter(tt.cmd)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCmdFilterCustomCommandName(t *testing.T) {
	c := &Config{CommandName: "agent"}

	// "agent" should be filtered out (it is the ophis command name).
	agentCmd := &cobra.Command{
		Use: "agent",
		Run: func(_ *cobra.Command, _ []string) {},
	}
	assert.True(t, c.cmdFilter(agentCmd))

	// "mcp" should NOT be filtered out (it is no longer the ophis command name).
	mcpCmd := &cobra.Command{
		Use: "mcp",
		Run: func(_ *cobra.Command, _ []string) {},
	}
	assert.False(t, c.cmdFilter(mcpCmd))

	// A normal command should pass through.
	normalCmd := &cobra.Command{
		Use: "status",
		Run: func(_ *cobra.Command, _ []string) {},
	}
	assert.False(t, c.cmdFilter(normalCmd))
}

func TestCommandNameDefault(t *testing.T) {
	// Empty CommandName defaults to "mcp".
	c := &Config{}
	assert.Equal(t, "mcp", c.commandName())

	// Explicit CommandName is used as-is.
	c = &Config{CommandName: "agent"}
	assert.Equal(t, "agent", c.commandName())

	// Nil receiver defaults to "mcp" (Command(nil) is a valid call).
	var nilConfig *Config
	assert.Equal(t, "mcp", nilConfig.commandName())
}
