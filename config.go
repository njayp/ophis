package ophis

import (
	"context"
	"log/slog"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/njayp/ophis/tools"
	"github.com/spf13/cobra"
)

// Config customizes MCP server behavior and command-to-tool conversion.
type Config struct {
	// GeneratorOptions controls command-to-tool conversion.
	// Default: Excludes non-runnable, hidden, and utility commands.
	GeneratorOptions []tools.GeneratorOption

	// SloggerOptions configures logging to stderr.
	// Default: Info level logging.
	SloggerOptions *slog.HandlerOptions

	// ServerOptions for the underlying MCP server.
	// See mark3labs/mcp-go documentation.
	ServerOptions []server.ServerOption

	// StdioOptions for stdio transport configuration.
	// See mark3labs/mcp-go documentation.
	StdioOptions []server.StdioOption
}

// tools returns the list of MCP tools generated from the root command.
func (c *Config) tools(rootCmd *cobra.Command) []tools.Controller {
	return tools.NewGenerator(c.GeneratorOptions...).FromRootCmd(rootCmd)
}

// setupSlogger configures logging to stderr to avoid corrupting stdio transport.
func (c *Config) setupSlogger() {
	handler := slog.NewTextHandler(os.Stderr, c.SloggerOptions)
	slog.SetDefault(slog.New(handler))
}

// serveStdio starts the MCP server with stdio transport.
func (c *Config) serveStdio(cmd *cobra.Command) error {
	c.setupSlogger()

	rootCmd := getRootCmd(cmd)
	appName := rootCmd.Name()
	version := rootCmd.Version
	slog.Info("creating MCP server", "app_name", appName, "app_version", version)

	srv := server.NewMCPServer(
		appName,
		version,
		c.ServerOptions...,
	)

	for _, ctrl := range c.tools(rootCmd) {
		slog.Debug("registering MCP tool", "tool_name", ctrl.Tool.Name)
		srv.AddTool(ctrl.Tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			slog.Info("MCP tool request received", "tool_name", ctrl.Tool.Name, "arguments", request.Params.Arguments)
			data, err := ctrl.Execute(ctx, request)
			return ctrl.Handle(ctx, request, data, err)
		})
	}

	return server.ServeStdio(srv, c.StdioOptions...)
}
