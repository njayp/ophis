package tools

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"
)

// Handler defines a function type for handling tool execution results.
// It takes the context, request, output data, and any error that occurred during execution,
// and returns an MCP CallToolResult or an error. Errors should be returned only if there is
// an issue with the handler itself, not with the tool execution.
type Handler func(context.Context, mcp.CallToolRequest, []byte, error) (*mcp.CallToolResult, error)

// WithHandler returns a GeneratorOption that sets a custom handler for processing command output.
// By default, the generator uses DefaultHandler which returns output as plain text.
func WithHandler(handler Handler) GeneratorOption {
	return func(g *Generator) {
		g.handler = handler
	}
}

// defaultHandler is the default handler that processes command output as plain text.
func defaultHandler(_ context.Context, request mcp.CallToolRequest, data []byte, err error) (*mcp.CallToolResult, error) {
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
