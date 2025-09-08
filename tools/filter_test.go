package tools

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// TestPathContains tests the pathContains function
func TestPathContains(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		phrase   string
		expected bool
	}{
		{"allows exact match", "get", "get", true},
		{"allows subcommand match", "user_get", "get", true},
		{"allows multi-word phrase match", "user_get_details", "user get", true},
		{"disallows non-matching command", "delete", "get", false},
		{"disallows partial non-matching phrase", "user_delete", "user get", false},
		{"allows multi-word phrase with extra words", "admin_user_get_info", "user get", true},
		{"allows multi-word phrase with extra words", "admin_user_get_info", "user info", false},
		{"disallows completely different command", "status", "user get", false},
		{"disallows single word match in multi-word phrase", "user", "user get", false},
		{"disallows when no words match", "create", "user get", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pathContains(tt.path, tt.phrase)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestExcludeFilter tests the Exclude filter function
func TestExcludeFilter(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		filter   Filter
		expected bool
	}{
		{"excludes if any phrase matches", "admin_user_test_info", Exclude([]string{"delete", "user test"}), false},
		{"allows if no phrases match", "admin_user_info", Exclude([]string{"delete", "user test"}), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: tt.path}
			result := tt.filter(cmd)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestAllowFilter tests the Allow filter function
func TestAllowFilter(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		filter   Filter
		expected bool
	}{
		{"allows if any phrase matches", "admin_user_get_info", Allow([]string{"delete", "user get"}), true},
		{"disallows if no phrases match", "admin_user_info", Allow([]string{"delete", "user get"}), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: tt.path}
			result := tt.filter(cmd)
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
