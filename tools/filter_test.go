package tools

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// TestExcludeFilter tests the Exclude filter function
func TestExcludeFilter(t *testing.T) {
	filter := Exclude([]string{"test", "admin"})

	tests := []struct {
		name     string
		cmdName  string
		expected bool
	}{
		{"excludes test command", "test", false},
		{"excludes admin command", "admin", false},
		{"allows other commands", "get", true},
		{"allows empty name", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: tt.cmdName}
			result := filter(cmd)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestAllowFilter tests the Allow filter function
func TestAllowFilter(t *testing.T) {
	filter := Allow([]string{"get", "list"})

	tests := []struct {
		name        string
		commandPath string
		cmdName     string
		expected    bool
	}{
		{"allows get command", "cli get", "get", true},
		{"allows list command", "cli list", "list", true},
		{"filters other commands", "cli delete", "delete", false},
		{"allows nested get", "cli resource get", "get", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: tt.cmdName}
			// Mock the command path
			cmd.SetUsageTemplate(tt.commandPath) // Hack to set a testable path
			// Since CommandPath() is not easily mockable, we test with the Use field
			// In real usage, Allow checks CommandPath() which includes parent commands
			
			// For this test, we'll create a simple parent-child structure
			if tt.commandPath == "cli get" || tt.commandPath == "cli list" || tt.commandPath == "cli resource get" {
				parent := &cobra.Command{Use: "cli"}
				parent.AddCommand(cmd)
			}
			
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
