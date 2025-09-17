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
	// Default: Excludes non-runnable, hidden, and utility commands.
	Filters []Filter

	// SloggerOptions configures logging to stderr.
	// Default: Info level logging.
	SloggerOptions *slog.HandlerOptions

	// ServerOptions for the underlying MCP server.
	ServerOptions *mcp.ServerOptions

	// Transport for stdio transport configuration.
	Transport mcp.Transport
}

func (c *Config) serveStdio(cmd *cobra.Command) error {
	// slog to stderr
	handler := slog.NewTextHandler(os.Stderr, c.SloggerOptions)
	slog.SetDefault(slog.New(handler))

	rootCmd := getRootCmd(cmd)
	appName := rootCmd.Name()
	version := rootCmd.Version
	slog.Info("creating MCP server", "app_name", appName, "app_version", version)

	server := mcp.NewServer(&mcp.Implementation{
		Name:    appName,
		Version: version,
	}, c.ServerOptions)

	for _, tool := range c.tools(rootCmd) {
		slog.Debug("registering MCP tool", "tool_name", tool.Name)
		mcp.AddTool(server, tool, bridge.Execute)
	}

	if c.Transport == nil {
		c.Transport = &mcp.StdioTransport{}
	}

	return server.Run(cmd.Context(), c.Transport)
}

func (c *Config) tools(rootCmd *cobra.Command) []*mcp.Tool {
	if c.Filters == nil {
		slog.Debug("using default filters")
		c.Filters = DefaultFilters()
	}

	slog.Debug("starting tool generation from root command", "root_cmd", rootCmd.Name())
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
	slog.Debug("created tool", "tool_name", tool.Name, "description", tool.Description)
	return append(tools, tool)
}
