package bridge

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/njayp/ophis/tools"
)

// registerTools recursively registers all Cobra commands as MCP tools
func (b *Manager) registerTools(tools []tools.Controller) {
	for _, tool := range tools {
		b.registerTool(tool)
	}
}

func (b *Manager) registerTool(ctrl tools.Controller) {
	slog.Debug("registering MCP tool", "tool_name", ctrl.Tool.Name)
	b.server.AddTool(ctrl.Tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		slog.Info("MCP tool request received", "tool_name", ctrl.Tool.Name, "arguments", request.Params.Arguments)
		data, err := ctrl.Execute(ctx, request)
		if err != nil {
			// Include output in error message if available
			output := string(data)
			slog.Error("command execution failed",
				"tool", ctrl.Tool.Name,
				"error", err,
				"output", output,
			)

			errMsg := fmt.Sprintf("command execution failed: %s", err.Error())
			if output != "" {
				errMsg += fmt.Sprintf("\nOutput: %s", output)
			}
			return mcp.NewToolResultError(errMsg), nil
		}

		return ctrl.Handle(ctx, request, data), nil
	})
}
