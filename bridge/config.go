package bridge

import (
	"log/slog"
	"os"

	"github.com/mark3labs/mcp-go/server"
	"github.com/njayp/ophis/tools"
	"github.com/spf13/cobra"
)

// Config holds configuration for the MCP command
type Config struct {
	AppName        string
	AppVersion     string
	RootCmd        *cobra.Command // The root command for the MCP server
	Generator      *tools.Generator
	SloggerOptions *slog.HandlerOptions
	ServerOptions  []server.ServerOption
}

// Tools returns the list of MCP tools generated from the root command.
// It uses the configured Generator if available, otherwise falls back to the default generator.
func (c *Config) Tools() []tools.Tool {
	if c.Generator != nil {
		return c.Generator.FromRootCmd(c.RootCmd)
	}

	return tools.FromRootCmd(c.RootCmd)
}

// setupSlogger makes a new slog.logger that writes to os.Stderr. Don't give the user
// the option to write to stdout, because that causes errors.
func (c *Config) setupSlogger() {
	// Create handler based on format preference
	handler := slog.NewTextHandler(os.Stderr, c.SloggerOptions)
	slog.SetDefault(slog.New(handler))
}
