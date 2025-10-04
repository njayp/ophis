package bridge

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildFlagArgs(t *testing.T) {
	tests := []struct {
		name     string
		flags    map[string]any
		expected []string
	}{
		{
			name:     "Empty flags",
			flags:    map[string]any{},
			expected: []string{},
		},
		{
			name: "Boolean flags",
			flags: map[string]any{
				"verbose": true,
				"quiet":   false,
				"debug":   true,
			},
			expected: []string{"--verbose", "--debug"},
		},
		{
			name: "String flags",
			flags: map[string]any{
				"output": "result.txt",
				"format": "json",
			},
			expected: []string{"--output", "result.txt", "--format", "json"},
		},
		{
			name: "Integer flags",
			flags: map[string]any{
				"count":   10,
				"timeout": 30,
			},
			expected: []string{"--count", "10", "--timeout", "30"},
		},
		{
			name: "Float flags",
			flags: map[string]any{
				"ratio":     0.75,
				"threshold": 1.5,
			},
			expected: []string{"--ratio", "0.75", "--threshold", "1.5"},
		},
		{
			name: "Array flags",
			flags: map[string]any{
				"include": []any{"*.go", "*.md"},
				"exclude": []any{"vendor"},
			},
			expected: []string{"--include", "*.go", "--include", "*.md", "--exclude", "vendor"},
		},
		{
			name: "Mixed types",
			flags: map[string]any{
				"verbose": true,
				"output":  "result.txt",
				"count":   5,
				"tags":    []any{"test", "debug"},
			},
			expected: []string{"--verbose", "--output", "result.txt", "--count", "5", "--tags", "test", "--tags", "debug"},
		},
		{
			name: "Nil values",
			flags: map[string]any{
				"flag1": nil,
				"flag2": "value",
				"flag3": nil,
			},
			expected: []string{"--flag2", "value"},
		},
		{
			name: "Empty flag name",
			flags: map[string]any{
				"":      "value",
				"valid": "value",
			},
			expected: []string{"--valid", "value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildFlagArgs(tt.flags)
			// Sort both slices for comparison since map iteration order is not guaranteed
			assert.ElementsMatch(t, tt.expected, result)
		})
	}
}

func TestBuildCommandArgs(t *testing.T) {
	tests := []struct {
		name         string
		commandName  string
		input        ToolInput
		expectedArgs []string
	}{
		{
			name:        "Simple command",
			commandName: "root_test",
			input: ToolInput{
				Flags: map[string]any{},
				Args:  []string{},
			},
			expectedArgs: []string{"test"},
		},
		{
			name:        "Nested command",
			commandName: "root_sub_command",
			input: ToolInput{
				Flags: map[string]any{},
				Args:  []string{},
			},
			expectedArgs: []string{"sub", "command"},
		},
		{
			name:        "Command with flags",
			commandName: "root_test",
			input: ToolInput{
				Flags: map[string]any{
					"verbose": true,
					"output":  "result.txt",
				},
				Args: []string{},
			},
			expectedArgs: []string{"test", "--verbose", "--output", "result.txt"},
		},
		{
			name:        "Command with arguments",
			commandName: "root_test",
			input: ToolInput{
				Flags: map[string]any{},
				Args:  []string{"file1.txt", "file2.txt"},
			},
			expectedArgs: []string{"test", "file1.txt", "file2.txt"},
		},
		{
			name:        "Command with flags and arguments",
			commandName: "root_deploy",
			input: ToolInput{
				Flags: map[string]any{
					"namespace": "production",
					"replicas":  3,
					"wait":      true,
				},
				Args: []string{"my-app", "v1.2.3"},
			},
			expectedArgs: []string{"deploy", "--namespace", "production", "--replicas", "3", "--wait", "my-app", "v1.2.3"},
		},
		{
			name:        "Complex nested command",
			commandName: "root_cluster_node_list",
			input: ToolInput{
				Flags: map[string]any{
					"output": "json",
					"label":  []any{"env=prod", "team=backend"},
				},
				Args: []string{},
			},
			expectedArgs: []string{"cluster", "node", "list", "--output", "json", "--label", "env=prod", "--label", "team=backend"},
		},
		{
			name:        "Command with quoted arguments",
			commandName: "root_exec",
			input: ToolInput{
				Flags: map[string]any{},
				Args:  []string{"argument with spaces", "another quoted arg", "normal"},
			},
			expectedArgs: []string{"exec", "argument with spaces", "another quoted arg", "normal"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildCommandArgs(tt.commandName, tt.input)

			// Extract command parts for comparison
			commandParts := len(result) - len(tt.expectedArgs)
			if commandParts >= 0 {
				// Compare command parts
				assert.Equal(t, tt.expectedArgs[:commandParts], result[:commandParts], "Command parts mismatch")
				// Compare flags and args (order might vary for flags)
				assert.ElementsMatch(t, tt.expectedArgs[commandParts:], result[commandParts:], "Flags/args mismatch")
			} else {
				assert.Equal(t, tt.expectedArgs, result)
			}
		})
	}
}
