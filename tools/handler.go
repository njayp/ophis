package tools

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
)

// Handler is a function type that processes the output from a command execution.
// It takes the MCP tool request and the command output data, and returns a tool result.
// Custom handlers can be used to format command output or handle special cases.
type Handler func(ctx context.Context, request mcp.CallToolRequest, data []byte) *mcp.CallToolResult

// WithHandler returns a GeneratorOption that sets a custom handler for processing command output.
// By default, the generator uses DefaultHandler which returns output as plain text.
func WithHandler(handler Handler) GeneratorOption {
	return func(g *Generator) {
		g.handler = handler
	}
}

// defaultHandler is the default handler that processes command output as plain text.
func defaultHandler(_ context.Context, _ mcp.CallToolRequest, data []byte) *mcp.CallToolResult {
	return mcp.NewToolResultText(string(data))
}
