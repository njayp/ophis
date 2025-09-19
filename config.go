package ophis

import (
	"log/slog"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/njayp/ophis/internal/bridge"
	"github.com/spf13/cobra"
)

// Config customizes MCP server behavior and command-to-tool conversion.
type Config struct {
	// Filters sets the filter for commands
	// Non-runnable, hidden, and utility commands are always excluded.
	Filters []Filter

	// Middleware gives the user flexibility to send metrics, timeouts, output, etc.
	Middleware *Middleware

	// SloggerOptions configures logging to stderr.
	// Default: Info level logging.
	SloggerOptions *slog.HandlerOptions

	// ServerOptions for the underlying MCP server.
	ServerOptions *mcp.ServerOptions

	// Transport for stdio transport configuration.
	Transport mcp.Transport
}

func (c *Config) serveStdio(cmd *cobra.Command) error {
	rootCmd := getRootCmd(cmd)
	name := rootCmd.Name()
	version := rootCmd.Version

	server := mcp.NewServer(&mcp.Implementation{
		Name:    name,
		Version: version,
	}, c.ServerOptions)

	for _, tool := range c.tools(rootCmd) {
		mcp.AddTool(server, tool, c.execute)
	}

	if c.Transport == nil {
		c.Transport = &mcp.StdioTransport{}
	}

	slog.Info("running MCP server", "name", name, "version", version)
	return server.Run(cmd.Context(), c.Transport)
}

// tools takes care of setup, and calls toolsRecursive
func (c *Config) tools(rootCmd *cobra.Command) []*mcp.Tool {
	// slog to stderr
	handler := slog.NewTextHandler(os.Stderr, c.SloggerOptions)
	slog.SetDefault(slog.New(handler))

	// add default filters
	c.Filters = append(c.Filters, defaultFilters()...)

	// get tools recursively
	tools := c.toolsRecursive(rootCmd, nil)
	slog.Info("tool generation completed", "total_tools", len(tools))
	return tools
}

func (c *Config) toolsRecursive(cmd *cobra.Command, tools []*mcp.Tool) []*mcp.Tool {
	if cmd == nil {
		return tools
	}

	// Register all subcommands
	for _, subCmd := range cmd.Commands() {
		tools = c.toolsRecursive(subCmd, tools)
	}

	// Apply all filters
	for _, filter := range c.Filters {
		if !filter(cmd) {
			return tools
		}
	}

	tool := bridge.CreateToolFromCmd(cmd)
	slog.Debug("created tool", "tool_name", tool.Name)
	return append(tools, tool)
}
