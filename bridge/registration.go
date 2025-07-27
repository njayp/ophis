package bridge

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/njayp/ophis/tools"
)

// registerTools recursively registers all Cobra commands as MCP tools
func (b *Manager) registerTools(tools []tools.Tool) {
	for _, tool := range tools {
		b.registerTool(tool)
	}
}

func (b *Manager) registerTool(t tools.Tool) {
	slog.Debug("registering MCP tool", "tool_name", t.Tool.Name)
	b.server.AddTool(t.Tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		slog.Info("MCP tool request received", "tool_name", t.Tool.Name, "arguments", request.Params.Arguments)
		data, err := b.executeCommand(ctx, t, request)
		if err != nil {
			// Include output in error message if available
			output := string(data)
			slog.Error("command execution failed",
				"tool", t.Tool.Name,
				"error", err,
				"output", output,
			)

			errMsg := fmt.Sprintf("command execution failed: %s", err.Error())
			if output != "" {
				errMsg += fmt.Sprintf("\nOutput: %s", output)
			}
			return mcp.NewToolResultError(errMsg), nil
		}

		return t.Handler(request, data), nil
	})
}
