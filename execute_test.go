package ophis

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildFlagArgs(t *testing.T) {
	tests := []struct {
		name  string
		flags map[string]any
		// names of flags that have json annotations
		jsonFlagNames []string
		expected      []string
	}{
		{
			name:          "Empty flags",
			flags:         map[string]any{},
			jsonFlagNames: []string{},
			expected:      []string{},
		},
		{
			name: "Boolean flags",
			flags: map[string]any{
				"verbose": true,
				"quiet":   false,
				"debug":   true,
			},
			jsonFlagNames: []string{},
			expected:      []string{"--verbose", "--debug"},
		},
		{
			name: "String flags",
			flags: map[string]any{
				"output": "result.txt",
				"format": "json",
			},
			jsonFlagNames: []string{},
			expected:      []string{"--output", "result.txt", "--format", "json"},
		},
		{
			name: "Integer flags",
			flags: map[string]any{
				"count":   10,
				"timeout": 30,
			},
			jsonFlagNames: []string{},
			expected:      []string{"--count", "10", "--timeout", "30"},
		},
		{
			name: "Float flags",
			flags: map[string]any{
				"ratio":     0.75,
				"threshold": 1.5,
			},
			jsonFlagNames: []string{},
			expected:      []string{"--ratio", "0.75", "--threshold", "1.5"},
		},
		{
			name: "Array flags",
			flags: map[string]any{
				"include": []any{"*.go", "*.md"},
				"exclude": []any{"vendor"},
			},
			jsonFlagNames: []string{},
			expected:      []string{"--include", "*.go", "--include", "*.md", "--exclude", "vendor"},
		},
		{
			name: "Mixed types",
			flags: map[string]any{
				"verbose": true,
				"output":  "result.txt",
				"count":   5,
				"tags":    []any{"test", "debug"},
			},
			jsonFlagNames: []string{},
			expected:      []string{"--verbose", "--output", "result.txt", "--count", "5", "--tags", "test", "--tags", "debug"},
		},
		{
			name: "Map flags (stringToString)",
			flags: map[string]any{
				"labels": map[string]any{
					"env":  "prod",
					"team": "backend",
				},
			},
			jsonFlagNames: []string{},
			expected:      []string{"--labels", "env=prod", "--labels", "team=backend"},
		},
		{
			name: "Empty map",
			flags: map[string]any{
				"labels": map[string]any{},
			},
			jsonFlagNames: []string{},
			expected:      []string{},
		},
		{
			name: "StringToInt map flags",
			flags: map[string]any{
				"ports": map[string]any{
					"http":  8080,
					"https": 8443,
				},
			},
			jsonFlagNames: []string{},
			expected:      []string{"--ports", "http=8080", "--ports", "https=8443"},
		},
		{
			name: "StringToInt64 map flags",
			flags: map[string]any{
				"sizes": map[string]any{
					"small":  int64(1024),
					"medium": int64(2048),
					"large":  int64(4096),
				},
			},
			jsonFlagNames: []string{},
			expected:      []string{"--sizes", "small=1024", "--sizes", "medium=2048", "--sizes", "large=4096"},
		},
		{
			name: "Nil values",
			flags: map[string]any{
				"flag1": nil,
				"flag2": "value",
				"flag3": nil,
			},
			jsonFlagNames: []string{},
			expected:      []string{"--flag2", "value"},
		},
		{
			name: "Empty flag name",
			flags: map[string]any{
				"":      "value",
				"valid": "value",
			},
			jsonFlagNames: []string{},
			expected:      []string{"--valid", "value"},
		},
		{
			name: "JSON schema object flag - simple",
			flags: map[string]any{
				"test-flag": map[string]any{
					"Name": "test",
				},
			},
			jsonFlagNames: []string{"test-flag"},
			expected:      []string{"--test-flag", `{"Name":"test"}`},
		},
		{
			name: "JSON schema object flag - complex nested",
			flags: map[string]any{
				"test-flag": map[string]any{
					"Name": "test",
					"Bar":  123,
					"Foobar": map[string]interface{}{
						"Baz": "nested",
					},
				},
			},
			jsonFlagNames: []string{"test-flag"},
			expected:      []string{"--test-flag", `{"Bar":123,"Foobar":{"Baz":"nested"},"Name":"test"}`},
		},
		{
			name: "JSON schema array flag",
			flags: map[string]any{
				"test-flag": []any{
					map[string]any{
						"Foo": "item1", "Bar": 1,
					},
					map[string]any{
						"Foo": "item2", "Bar": 2,
					},
				},
			},
			jsonFlagNames: []string{"test-flag"},
			expected:      []string{"--test-flag", `[{"Bar":1,"Foo":"item1"},{"Bar":2,"Foo":"item2"}]`},
		},
		{
			name: "mixed regular and JSON flags",
			flags: map[string]any{
				"test-flag": map[string]any{
					"Name": "test",
					"Bar":  123,
					"Foobar": map[string]interface{}{
						"Baz": "nested",
					},
				},
				"not-json": "some-value",
			},
			jsonFlagNames: []string{"test-flag"},
			expected: []string{
				"--not-json", "some-value",
				"--test-flag", `{"Bar":123,"Foobar":{"Baz":"nested"},"Name":"test"}`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testToolName := "foo"
			metadata := make(FlagMetadataByFlagName)
			for name := range tt.flags {
				if slices.Contains(tt.jsonFlagNames, name) {
					metadata[name] = FlagMetadata{HasJSONSchema: true}
				} else {
					metadata[name] = FlagMetadata{HasJSONSchema: false}
				}
			}
			globalFlagRegistry.Register(testToolName, metadata)
			result := buildFlagArgs(testToolName, tt.flags)
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
			name:        "Command with map flags",
			commandName: "root_deploy",
			input: ToolInput{
				Flags: map[string]any{
					"labels": map[string]any{
						"env":     "production",
						"version": "v1.2.3",
					},
					"wait": true,
				},
				Args: []string{"my-app"},
			},
			expectedArgs: []string{"deploy", "--labels", "env=production", "--labels", "version=v1.2.3", "--wait", "my-app"},
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
