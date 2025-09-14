package tools

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"
)

// Handler processes tool execution results into MCP responses.
// Returns an error only for handler failures, not tool execution errors.
type Handler func(context.Context, mcp.CallToolRequest, []byte, error) (*mcp.CallToolResult, error)

// WithHandler sets a custom output handler.
// Default: Returns output as plain text.
func WithHandler(handler Handler) GeneratorOption {
	return func(g *Generator) {
		g.handler = handler
	}
}

// DefaultHandler returns command output as plain text with error details.
func DefaultHandler(_ context.Context, request mcp.CallToolRequest, data []byte, err error) (*mcp.CallToolResult, error) {
	output := string(data)
	if err != nil {
		slog.Error("command execution failed",
			"tool", request.Method,
			"error", err,
			"output", output,
		)

		// Include output in error message if available
		errMsg := fmt.Sprintf("command execution failed: %s", err.Error())
		if output != "" {
			errMsg += fmt.Sprintf("\nOutput: %s", output)
		}
		return mcp.NewToolResultError(errMsg), nil
	}

	return mcp.NewToolResultText(output), nil
}
