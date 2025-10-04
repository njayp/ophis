package bridge

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// buildCommandTree creates a command tree from a list of command names.
// The first command becomes the root, and subsequent commands are nested.
func buildCommandTree(names ...string) *cobra.Command {
	if len(names) == 0 {
		return nil
	}

	root := &cobra.Command{Use: names[0]}
	parent := root

	for _, name := range names[1:] {
		child := &cobra.Command{
			Use: name,
			Run: func(_ *cobra.Command, _ []string) {},
		}

		parent.AddCommand(child)
		parent = child
	}

	return parent
}

func TestCmdSelect(t *testing.T) {
	tests := []struct {
		name         string
		allowPhrases []string
		commandNames []string
		expected     bool
	}{
		{
			name:         "matches single phrase",
			allowPhrases: []string{"get"},
			commandNames: []string{"kubectl", "get", "pods"},
			expected:     true,
		},
		{
			name:         "matches one of multiple phrases",
			allowPhrases: []string{"get", "list"},
			commandNames: []string{"helm", "list"},
			expected:     true,
		},
		{
			name:         "matches exact command name",
			allowPhrases: []string{"kubectl get"},
			commandNames: []string{"kubectl", "get", "pods"},
			expected:     true,
		},
		{
			name:         "does not match when phrase absent",
			allowPhrases: []string{"delete"},
			commandNames: []string{"kubectl", "get", "pods"},
			expected:     false,
		},
		{
			name:         "does not match any of multiple phrases",
			allowPhrases: []string{"delete", "remove"},
			commandNames: []string{"kubectl", "get", "pods"},
			expected:     false,
		},
		{
			name:         "partial match in middle of path",
			allowPhrases: []string{"admin"},
			commandNames: []string{"cli", "admin", "user"},
			expected:     true,
		},
		{
			name:         "empty phrases list rejects everything",
			allowPhrases: []string{},
			commandNames: []string{"kubectl", "get", "pods"},
			expected:     false,
		},
		{
			name:         "case sensitive matching",
			allowPhrases: []string{"Get"},
			commandNames: []string{"kubectl", "get", "pods"},
			expected:     false,
		},
		{
			name:         "matches substring in command name",
			allowPhrases: []string{"pod"},
			commandNames: []string{"kubectl", "get", "pods"},
			expected:     true,
		},
		{
			name:         "rejects help command",
			allowPhrases: []string{"get"},
			commandNames: []string{"kubectl", "get", "help"},
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := buildCommandTree(tt.commandNames...)
			selector := Selector{
				CmdSelector: CmdContains(tt.allowPhrases...),
			}
			result := selector.cmdSelect(cmd)
			assert.Equal(t, tt.expected, result)
		})
	}
}
