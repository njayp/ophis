package tools

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestExcludeFilter(t *testing.T) {
	filter := Exclude([]string{"root mcp", "admin"})

	tests := []struct {
		name     string
		cmdName  string
		expected bool
	}{
		{"filters excluded command 'root mcp'", "mcp", false},
		{"filters excluded command 'admin'", "admin", false},
		{"allows non-excluded command 'get'", "get", true},
		{"allows non-excluded command 'list'", "list", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := &cobra.Command{Use: "root"}
			cmd := &cobra.Command{
				Use: tt.cmdName,
			}
			root.AddCommand(cmd)
			result := filter(cmd)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAllowFilter(t *testing.T) {
	filter := Allow([]string{"root get", "admin"})

	tests := []struct {
		name     string
		cmdName  string
		expected bool
	}{
		{"allows included command 'root get'", "get", true},
		{"allows included command 'admin'", "admin", true},
		{"filters non-included command 'list'", "list", false},
		{"filters non-included command 'mcp'", "mcp", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := &cobra.Command{Use: "root"}
			cmd := &cobra.Command{
				Use: tt.cmdName,
			}
			root.AddCommand(cmd)
			result := filter(cmd)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestHiddenFilter tests the Hidden filter function
func TestHiddenFilter(t *testing.T) {
	filter := Hidden()

	tests := []struct {
		name     string
		hidden   bool
		expected bool
	}{
		{"filters hidden command", true, false},
		{"allows visible command", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use:    "test",
				Hidden: tt.hidden,
			}
			result := filter(cmd)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFilterChaining tests that multiple filters work together
func TestFilterChaining(t *testing.T) {
	gen := NewGenerator(
		WithFilters(
			Hidden(),
			Exclude([]string{"admin"}),
		),
	)

	tests := []struct {
		name     string
		cmd      *cobra.Command
		expected bool // Should the command be included?
	}{
		{
			name: "visible non-admin command passes",
			cmd: &cobra.Command{
				Use:    "get",
				Hidden: false,
			},
			expected: true,
		},
		{
			name: "hidden command filtered",
			cmd: &cobra.Command{
				Use:    "get",
				Hidden: true,
			},
			expected: false,
		},
		{
			name: "admin command filtered",
			cmd: &cobra.Command{
				Use:    "admin",
				Hidden: false,
			},
			expected: false,
		},
		{
			name: "hidden admin command filtered",
			cmd: &cobra.Command{
				Use:    "admin",
				Hidden: true,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Apply all filters
			passed := true
			for _, filter := range gen.filters {
				if !filter(tt.cmd) {
					passed = false
					break
				}
			}
			assert.Equal(t, tt.expected, passed)
		})
	}
}

// TestAddFilter tests adding filters to existing ones
func TestAddFilter(t *testing.T) {
	customFilter := func(cmd *cobra.Command) bool {
		return cmd.Use != "blocked"
	}

	gen := NewGenerator(AddFilter(customFilter))

	// Should have default filters plus the custom one
	assert.Len(t, gen.filters, 3) // 2 defaults + 1 custom

	// Test that all filters are applied
	blockedCmd := &cobra.Command{Use: "blocked"}
	mcpCmd := &cobra.Command{Use: "mcp"} // Should be blocked by default
	normalCmd := &cobra.Command{Use: "normal"}

	// Test the filter behavior
	root := &cobra.Command{Use: "root"}
	root.AddCommand(blockedCmd, mcpCmd, normalCmd)

	// Only add Run functions to commands we want to see in results
	normalCmd.Run = func(_ *cobra.Command, _ []string) {}
	blockedCmd.Run = func(_ *cobra.Command, _ []string) {}
	mcpCmd.Run = func(_ *cobra.Command, _ []string) {}

	tools := gen.FromRootCmd(root)

	// Should only have the normal command
	assert.Len(t, tools, 1)
	assert.Equal(t, "root_normal", tools[0].Tool.Name)
}
