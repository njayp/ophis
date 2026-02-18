package manager

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeriveServerName(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		expectedName string
	}{
		{
			name:         "simple executable",
			path:         "/usr/local/bin/myapp",
			expectedName: "myapp",
		},
		{
			name:         "executable with extension",
			path:         "/usr/local/bin/myapp.exe",
			expectedName: "myapp",
		},
		{
			name:         "nested path",
			path:         "/home/user/dev/project/bin/mycli",
			expectedName: "mycli",
		},
		{
			name:         "current directory",
			path:         "./myapp",
			expectedName: "myapp",
		},
		{
			name:         "just filename",
			path:         "kubectl",
			expectedName: "kubectl",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DeriveServerName(tt.path)
			assert.Equal(t, tt.expectedName, result)
		})
	}
}

func TestGetCmdPath(t *testing.T) {
	tests := []struct {
		name         string
		cmdPath      string
		commandName  string
		expectedPath []string
		expectError  bool
	}{
		{
			name:         "root mcp command",
			cmdPath:      "myapp mcp",
			commandName:  "mcp",
			expectedPath: []string{"mcp"},
			expectError:  false,
		},
		{
			name:         "nested mcp command",
			cmdPath:      "myapp alpha mcp",
			commandName:  "mcp",
			expectedPath: []string{"alpha", "mcp"},
			expectError:  false,
		},
		{
			name:         "deeply nested mcp command",
			cmdPath:      "myapp alpha beta mcp start",
			commandName:  "mcp",
			expectedPath: []string{"alpha", "beta", "mcp"},
			expectError:  false,
		},
		{
			name:         "no mcp in path",
			cmdPath:      "myapp start",
			commandName:  "mcp",
			expectedPath: nil,
			expectError:  true,
		},
		{
			name:         "mcp as root returns empty slice",
			cmdPath:      "mcp start",
			commandName:  "mcp",
			expectedPath: []string{},
			expectError:  false,
		},
		// Custom command name tests
		{
			name:         "renamed to agent",
			cmdPath:      "devenv agent claude enable",
			commandName:  "agent",
			expectedPath: []string{"agent"},
			expectError:  false,
		},
		{
			name:         "renamed to agent nested",
			cmdPath:      "myapp sub agent vscode enable",
			commandName:  "agent",
			expectedPath: []string{"sub", "agent"},
			expectError:  false,
		},
		{
			name:         "renamed but searching for mcp fails",
			cmdPath:      "devenv agent claude enable",
			commandName:  "mcp",
			expectedPath: nil,
			expectError:  true,
		},
		{
			name:         "custom name as root returns empty slice",
			cmdPath:      "agent start",
			commandName:  "agent",
			expectedPath: []string{},
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build a command tree that matches the path
			cmd := buildMockCommandTree(tt.cmdPath)

			result, err := GetCmdPath(cmd, tt.commandName)

			if tt.expectError {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedPath, result)
			}
		})
	}
}

// buildMockCommandTree creates a cobra command tree from a space-separated path.
// The last command in the path is returned.
func buildMockCommandTree(path string) *cobra.Command {
	parts := strings.Fields(path)
	if len(parts) == 0 {
		return nil
	}

	root := &cobra.Command{Use: parts[0]}
	parent := root

	for i := 1; i < len(parts); i++ {
		cmd := &cobra.Command{Use: parts[i]}
		parent.AddCommand(cmd)
		parent = cmd
	}

	return parent
}
