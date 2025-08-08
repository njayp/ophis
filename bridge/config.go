package bridge

import (
	"log/slog"
	"os"

	"github.com/mark3labs/mcp-go/server"
	"github.com/njayp/ophis/tools"
	"github.com/spf13/cobra"
)

// Config holds configuration for creating an MCP server bridge from a Cobra CLI application.
// It defines how the CLI commands are exposed as MCP tools and how the server behaves.
//
// Example usage:
//
//	config := &bridge.Config{
//		RootCmd:    rootCmd,
//		Generator: tools.NewGenerator(
//			tools.WithFilters(tools.Allow([]string{"get", "list"})),
//			tools.WithHandler(customHandler),
//		),
//		SloggerOptions: &slog.HandlerOptions{
//			Level: slog.LevelDebug,
//		},
//	}
type Config struct {
	// RootCmd is the root Cobra command whose subcommands will be exposed as MCP tools.
	// Required: This is the entry point for discovering available commands.
	// Only commands with Run or RunE functions will be converted to tools.
	RootCmd *cobra.Command

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

// Tools returns the list of MCP tools generated from the root command.
//
// If a custom Generator is configured, it uses that to convert commands.
// Otherwise, it falls back to the default generator which:
//   - Excludes hidden commands
//   - Excludes "mcp", "help", and "completion" commands
//   - Returns command output as plain text
func (c *Config) Tools() []tools.Controller {
	if c.Generator != nil {
		return c.Generator.FromRootCmd(c.RootCmd)
	}

	return tools.FromRootCmd(c.RootCmd)
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
