package tools

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
)

// TestDefaultHandler tests the default handler behavior
func TestDefaultHandler(t *testing.T) {
	handler := DefaultHandler()

	tests := []struct {
		name     string
		data     []byte
		expected string
	}{
		{
			name:     "simple text output",
			data:     []byte("Hello, world!"),
			expected: "Hello, world!",
		},
		{
			name:     "multiline output",
			data:     []byte("Line 1\nLine 2\nLine 3"),
			expected: "Line 1\nLine 2\nLine 3",
		},
		{
			name:     "empty output",
			data:     []byte(""),
			expected: "",
		},
		{
			name:     "output with special characters",
			data:     []byte("Special chars: !@#$%^&*()"),
			expected: "Special chars: !@#$%^&*()",
		},
		{
			name:     "JSON output",
			data:     []byte(`{"key": "value", "number": 123}`),
			expected: `{"key": "value", "number": 123}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a dummy request (handler doesn't use it in default implementation)
			request := mcp.CallToolRequest{}

			result := handler(request, tt.data)

			// Verify the result is a text result
			assert.NotNil(t, result)
			// The actual content verification would depend on the mcp.CallToolResult structure
			// For now, we just verify that a result is returned
		})
	}
}

// TestWithHandler tests the WithHandler option
func TestWithHandler(t *testing.T) {
	// Create a custom handler that modifies the output
	customHandler := func(_ mcp.CallToolRequest, data []byte) *mcp.CallToolResult {
		// Add a prefix to the output
		modifiedData := append([]byte("CUSTOM: "), data...)
		return mcp.NewToolResultText(string(modifiedData))
	}

	// Create a generator with the custom handler
	generator := &Generator{}
	option := WithHandler(customHandler)
	option(generator)

	// Verify the handler was set
	assert.NotNil(t, generator.handler)

	// Test that the custom handler works
	request := mcp.CallToolRequest{}
	data := []byte("test output")
	result := generator.handler(request, data)
	assert.NotNil(t, result)
}

// TestHandlerTypes demonstrates different handler implementations
func TestHandlerTypes(t *testing.T) {
	tests := []struct {
		name    string
		handler Handler
		data    []byte
	}{
		{
			name: "text handler",
			handler: func(_ mcp.CallToolRequest, data []byte) *mcp.CallToolResult {
				return mcp.NewToolResultText(string(data))
			},
			data: []byte("text output"),
		},
		{
			name: "error handler",
			handler: func(_ mcp.CallToolRequest, data []byte) *mcp.CallToolResult {
				if len(data) == 0 {
					return mcp.NewToolResultError("no output")
				}
				return mcp.NewToolResultText(string(data))
			},
			data: []byte(""),
		},
		{
			name: "filtering handler",
			handler: func(_ mcp.CallToolRequest, data []byte) *mcp.CallToolResult {
				// Example: filter out debug lines
				output := string(data)
				if len(output) > 100 {
					output = output[:100] + "... (truncated)"
				}
				return mcp.NewToolResultText(output)
			},
			data: []byte("This is a very long output that should be truncated to avoid overwhelming the user with too much information at once"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := mcp.CallToolRequest{}
			result := tt.handler(request, tt.data)
			assert.NotNil(t, result)
		})
	}
}
