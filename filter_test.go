package ophis

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestExcludeFilter(t *testing.T) {
	filter := ExcludeFilter([]string{"root mcp", "admin"})

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
	filter := AllowFilter([]string{"root get", "admin"})

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
	filter := HiddenFilter()

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

func TestRunsFilter(t *testing.T) {
	filter := RunsFilter()

	tests := []struct {
		name     string
		cmd      *cobra.Command
		expected bool
	}{
		{"filters command without Run", &cobra.Command{Use: "no-run"}, false},
		{"allows command with Run", &cobra.Command{Use: "has-run", Run: func(*cobra.Command, []string) {}}, true},
		{"allows command with RunE", &cobra.Command{Use: "has-runE", RunE: func(*cobra.Command, []string) error { return nil }}, true},
		{"allows command with PreRun", &cobra.Command{Use: "has-preRun", PreRun: func(*cobra.Command, []string) {}}, true},
		{"allows command with PreRunE", &cobra.Command{Use: "has-preRunE", PreRunE: func(*cobra.Command, []string) error { return nil }}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filter(tt.cmd)
			assert.Equal(t, tt.expected, result)
		})
	}
}
