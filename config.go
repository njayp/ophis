package ophis

import (
	"log/slog"

	"github.com/mark3labs/mcp-go/server"
	"github.com/njayp/ophis/internal/bridge"
	"github.com/njayp/ophis/tools"
	"github.com/spf13/cobra"
)

// Config provides user options for creating MCP commands for the MCP server bridge.
// It defines how the CLI commands are exposed as MCP tools and how the server behaves.
//
// Example usage:
//
//	config := &bridge.Config{
//		Generator: tools.NewGenerator(
//			tools.WithFilters(tools.Allow([]string{"get", "list"})),
//			tools.WithHandler(customHandler),
//		),
//		SloggerOptions: &slog.HandlerOptions{
//			Level: slog.LevelDebug,
//		},
//	}
type Config struct {
	// Generator controls how Cobra commands are converted to MCP tools.
	// Optional: If nil, a default generator will be used that:
	//   - Excludes hidden commands
	//   - Excludes "mcp", "help", and "completion" commands
	//   - Returns command output as plain text
	//
	// Example:
	//   config.Generator = tools.NewGenerator(
	//       tools.WithFilters(tools.Allow([]string{"get", "list"})),
	//       tools.WithHandler(customHandler),
	//   )
	Generator *tools.Generator

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
}

func (c *Config) bridgeConfig(cmd *cobra.Command) *bridge.Config {
	rootCmd := cmd.Parent().Parent()

	return &bridge.Config{
		RootCmd:        rootCmd,
		Generator:      c.Generator,
		SloggerOptions: c.SloggerOptions,
		ServerOptions:  c.ServerOptions,
	}
}
