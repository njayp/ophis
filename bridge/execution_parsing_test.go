package bridge

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseArgumentString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
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
			name:     "simple arguments",
			input:    "arg1 arg2 arg3",
			expected: []string{"arg1", "arg2", "arg3"},
		},
		{
			name:     "arguments with extra spaces",
			input:    "  arg1   arg2  arg3  ",
			expected: []string{"arg1", "arg2", "arg3"},
		},
		{
			name:     "quoted string with spaces",
			input:    `arg1 "hello world" arg3`,
			expected: []string{"arg1", "hello world", "arg3"},
		},
		{
			name:     "single quoted string",
			input:    `arg1 'hello world' arg3`,
			expected: []string{"arg1", "hello world", "arg3"},
		},
		{
			name:     "escaped spaces",
			input:    `arg1 hello\ world arg3`,
			expected: []string{"arg1", "hello world", "arg3"},
		},
		{
			name:     "mixed quotes",
			input:    `"double quoted" 'single quoted' unquoted`,
			expected: []string{"double quoted", "single quoted", "unquoted"},
		},
		{
			name:     "nested quotes",
			input:    `"outer 'inner' quotes"`,
			expected: []string{"outer 'inner' quotes"},
		},
		{
			name:     "escaped quotes",
			input:    `"escaped \" quote"`,
			expected: []string{`escaped " quote`},
		},
		{
			name:     "file paths with spaces",
			input:    `--file="/path/to/some file.txt" --output="another file.log"`,
			expected: []string{"--file=/path/to/some file.txt", "--output=another file.log"},
		},
		{
			name:     "complex real-world example",
			input:    `cmd -v --format="json pretty" --input='/data/My Documents/file.txt' --tags=prod,staging output.json`,
			expected: []string{"cmd", "-v", "--format=json pretty", "--input=/data/My Documents/file.txt", "--tags=prod,staging", "output.json"},
		},
		{
			name:     "empty quotes",
			input:    `arg1 "" arg3`,
			expected: []string{"arg1", "", "arg3"},
		},
		{
			name:     "backslash at end",
			input:    `arg1 arg2\`,
			expected: []string{"arg1", "arg2\\"},
		},
		{
			name:     "unterminated double quote",
			input:    `arg1 "unterminated`,
			expected: []string{"arg1", `"unterminated`}, // Falls back to simple splitting
		},
		{
			name:     "unterminated single quote",
			input:    `arg1 'unterminated`,
			expected: []string{"arg1", "'unterminated"}, // Falls back to simple splitting
		},
		{
			name:     "special shell characters",
			input:    `echo "Hello $USER" > output.txt`,
			expected: []string{"echo", "Hello $USER", ">", "output.txt"},
		},
		{
			name:     "glob patterns",
			input:    `ls *.txt "file with space.txt"`,
			expected: []string{"ls", "*.txt", "file with space.txt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseArgumentString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
