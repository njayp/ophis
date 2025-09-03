package bridge

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/njayp/ophis/tools"
)

// manager manages the bridge between a Cobra CLI application and an MCP server.
// It handles tool registration, command execution, and server lifecycle.
//
// manager instances should be created using the New function rather than
// direct struct initialization to ensure proper validation and setup.
type manager struct {
	server *server.MCPServer // The underlying MCP server instance
}

// Run creates and starts an MCP server bridge using the provided configuration.
// Returns an error if:
//   - config is nil
//   - config.RootCmd is nil
func Run(config *Config) error {
	if config == nil {
		return fmt.Errorf("configuration cannot be nil: must provide a Config struct with a RootCmd")
	}

	if config.RootCmd == nil {
		return fmt.Errorf("root command cannot be nil: config.RootCmd is required to register tools")
	}

	config.setupSlogger()

	appName := config.RootCmd.Name()
	version := config.RootCmd.Version
	slog.Info("creating MCP server", "app_name", appName, "app_version", version)

	b := &manager{
		server: server.NewMCPServer(
			appName,
			version,
			config.ServerOptions...,
		),
	}

	b.registerTools(config.Tools())
	return server.ServeStdio(b.server)
}

// registerTools recursively registers all Cobra commands as MCP tools
func (b *manager) registerTools(tools []tools.Controller) {
	for _, tool := range tools {
		b.registerTool(tool)
	}
}

func (b *manager) registerTool(ctrl tools.Controller) {
	slog.Debug("registering MCP tool", "tool_name", ctrl.Tool.Name)
	b.server.AddTool(ctrl.Tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		slog.Info("MCP tool request received", "tool_name", ctrl.Tool.Name, "arguments", request.Params.Arguments)
		data, err := ctrl.Execute(ctx, request)
		return ctrl.Handle(ctx, request, data, err)
	})
}
