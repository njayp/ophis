package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestParseArgumentString tests the shell-like argument parsing
func TestParseArgumentString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "simple space-separated args",
			input:    "foo bar baz",
			expected: []string{"foo", "bar", "baz"},
		},
		{
			name:     "double quoted string",
			input:    `foo "bar baz" qux`,
			expected: []string{"foo", "bar baz", "qux"},
		},
		{
			name:     "single quoted string",
			input:    `foo 'bar baz' qux`,
			expected: []string{"foo", "bar baz", "qux"},
		},
		{
			name:     "escaped space",
			input:    `foo bar\ baz`,
			expected: []string{"foo", "bar baz"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "whitespace only",
			input:    "   ",
			expected: nil,
		},
		{
			name:     "complex quoting",
			input:    `cmd --flag="value with spaces" 'another value'`,
			expected: []string{"cmd", "--flag=value with spaces", "another value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseArgumentString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestBuildFlagArgs tests flag argument construction
func TestBuildFlagArgs(t *testing.T) {
	tests := []struct {
		name     string
		flagMap  map[string]any
		expected []string
	}{
		{
			name: "boolean true flag",
			flagMap: map[string]any{
				"verbose": true,
			},
			expected: []string{"--verbose"},
		},
		{
			name: "boolean false flag",
			flagMap: map[string]any{
				"verbose": false,
			},
			expected: nil,
		},
		{
			name: "string flag",
			flagMap: map[string]any{
				"output": "json",
			},
			expected: []string{"--output", "json"},
		},
		{
			name: "integer flag",
			flagMap: map[string]any{
				"count": 42,
			},
			expected: []string{"--count", "42"},
		},
		{
			name: "multiple flags",
			flagMap: map[string]any{
				"verbose": true,
				"output":  "json",
				"quiet":   false,
			},
			// Order may vary, so we check length and contents
			expected: []string{"--verbose", "--output", "json"},
		},
		{
			name:     "empty flag map",
			flagMap:  map[string]any{},
			expected: nil,
		},
		{
			name: "nil value flag",
			flagMap: map[string]any{
				"flag": nil,
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildFlagArgs(tt.flagMap)

			// For multiple flags test, check elements match regardless of order
			if tt.name == "multiple flags" {
				assert.ElementsMatch(t, tt.expected, result)
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
