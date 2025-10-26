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
			result := cmdFilter(tt.cmd)
			assert.Equal(t, tt.expected, result)
		})
	}
}
