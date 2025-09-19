package ophis

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/njayp/ophis/internal/bridge"
)

// Middleware gives the user flexibility to send metrics, limit resources, etc.
type Middleware struct {
	// Add timeouts or cancel using the context
	PreRun  func(context.Context, *mcp.CallToolRequest, bridge.CmdToolInput) (context.Context, *mcp.CallToolRequest, bridge.CmdToolInput)
	PostRun func(context.Context, *mcp.CallToolRequest, bridge.CmdToolInput, *mcp.CallToolResult, bridge.CmdToolOutput, error) (*mcp.CallToolResult, bridge.CmdToolOutput, error)
}

func (c *Config) execute(ctx context.Context, request *mcp.CallToolRequest, input bridge.CmdToolInput) (result *mcp.CallToolResult, output bridge.CmdToolOutput, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()

	if c.Middleware == nil {
		return bridge.Execute(ctx, request, input)
	}

	if c.Middleware.PreRun != nil {
		ctx, request, input = c.Middleware.PreRun(ctx, request, input)
	}

	result, output, err = bridge.Execute(ctx, request, input)

	if c.Middleware.PostRun != nil {
		result, output, err = c.Middleware.PostRun(ctx, request, input, result, output, err)
	}

	return
}
