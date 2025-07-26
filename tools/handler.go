package tools

import "github.com/mark3labs/mcp-go/mcp"

type Handler func(request mcp.CallToolRequest, data []byte) *mcp.CallToolResult

func (g *Generator) WithHandler(handler Handler) {
	g.handler = handler
}

func defaultHandler(_ mcp.CallToolRequest, data []byte) *mcp.CallToolResult {
	return mcp.NewToolResultText(string(data))
}
