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

// Config provides user options for creating MCP commands for the MCP server bridge.
// It defines how the CLI commands are exposed as MCP tools and how the server behaves.
//
// Example usage:
//
//	config := &bridge.Config{
//		GeneratorOptions: []tools.GeneratorOption{
//			tools.WithFilters(tools.Allow([]string{"get", "list"})),
//			tools.WithHandler(customHandler),
//		},
//		SloggerOptions: &slog.HandlerOptions{
//			Level: slog.LevelDebug,
//		},
//	}
type Config struct {
	// Generator controls how Cobra commands are converted to MCP tools.
	// Optional: If nil, a default generator will be used that:
	//   - Excludes commands without a Run or PreRun function
	//   - Excludes hidden commands
	//   - Excludes "mcp", "help", and "completion" commands
	//   - Returns command output as plain text
	GeneratorOptions []tools.GeneratorOption

	// SloggerOptions configures the structured logger used by the MCP server.
	// Optional: If nil, default options will be used.
	// The logger always writes to stderr to avoid interfering with stdio transport.
	//
	// Example:
	//   config.SloggerOptions = &slog.HandlerOptions{
	//       Level: slog.LevelDebug,  // Enable debug logging
	//   }
	SloggerOptions *slog.HandlerOptions

	// ServerOptions provides additional options for the underlying MCP server.
	// Optional: These are passed directly to the mark3labs/mcp-go server.
	// The bridge always adds server.WithRecovery() to handle panics gracefully.
	//
	// Consult the mark3labs/mcp-go documentation for available server options.
	ServerOptions []server.ServerOption

	// StdioOptions provides additional options for stdio transport.
	// Optional: These are passed directly to server.ServeStdio.
	// The bridge always adds server.WithStdioLogging() to log stdio transport events.
	//
	// Consult the mark3labs/mcp-go documentation for available stdio options.
	StdioOptions []server.StdioOption
}

// tools returns the list of MCP tools generated from the root command.
func (c *Config) tools(rootCmd *cobra.Command) []tools.Controller {
	return tools.NewGenerator(c.GeneratorOptions...).FromRootCmd(rootCmd)
}

// setupSlogger configures the structured logger for the MCP server.
//
// The logger always writes to stderr to avoid interfering with the stdio
// transport used for MCP communication. Writing logs to stdout would corrupt
// the MCP protocol messages.
func (c *Config) setupSlogger() {
	handler := slog.NewTextHandler(os.Stderr, c.SloggerOptions)
	slog.SetDefault(slog.New(handler))
}

// serveStdio creates and starts an MCP server using stdio transport.
//
// It sets up the logger, registers tools generated from the provided Cobra
// command, and starts serving requests over stdio.
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
