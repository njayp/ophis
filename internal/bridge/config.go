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
type Config struct {
	// RootCmd is the root Cobra command whose subcommands will be exposed as MCP tools.
	// Required: This is the entry point for discovering available commands.
	RootCmd *cobra.Command

	// Generator controls how Cobra commands are converted to MCP tools.
	GeneratorOptions []tools.GeneratorOption

	// SloggerOptions configures the structured logger used by the MCP server.
	SloggerOptions *slog.HandlerOptions

	// ServerOptions provides additional options for the underlying MCP server.
	ServerOptions []server.ServerOption

	// StreamOptions provides options for the HTTP stream transport.
	StreamOptions []server.StreamableHTTPOption
}

// Tools returns the list of MCP tools generated from the root command.
func (c *Config) Tools() []tools.Controller {
	return tools.NewGenerator(c.GeneratorOptions...).FromRootCmd(c.RootCmd)
}

// SetupSlogger configures the structured logger for the MCP server.
//
// The logger always writes to stderr to avoid interfering with the stdio
// transport used for MCP communication. Writing logs to stdout would corrupt
// the MCP protocol messages.
func (c *Config) SetupSlogger() {
	handler := slog.NewTextHandler(os.Stderr, c.SloggerOptions)
	slog.SetDefault(slog.New(handler))
}
