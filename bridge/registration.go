package bridge

import (
	"context"

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
	b.logger.Debug("Registering MCP tool", "tool_name", t.Tool.Name)
	b.server.AddTool(t.Tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		b.logger.Info("MCP tool request received", "tool_name", t.Tool.Name, "arguments", request.Params.Arguments)
		result := b.executeCommand(ctx, t, request)
		// TODO figure out what err is used for
		return result, nil
	})
}
