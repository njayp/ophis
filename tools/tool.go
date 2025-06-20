package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

// Constants for MCP parameter names and error messages
const (
	MCPCommandName = "mcp"
	// PositionalArgsParam is the parameter name for positional arguments
	PositionalArgsParam = "args"
	FlagsParam          = "flags"
)

// Tool represents an MCP tool with its associated metadata
type Tool struct {
	Tool mcp.Tool `json:"tool"`
	// more fields will likely be necessary in the future
}
