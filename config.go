package ophis

import (
	"log/slog"
	"os"

	"github.com/google/jsonschema-go/jsonschema"
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

	// ForOptions for schema creation
	ForOptions *jsonschema.ForOptions

	// ServerOptions for the underlying MCP server.
	ServerOptions *mcp.ServerOptions

	// Transport for stdio transport configuration.
	Transport mcp.Transport
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

	srv := mcp.NewServer(&mcp.Implementation{
		Name:    appName,
		Version: version,
	}, c.ServerOptions)

	for _, tool := range c.tools(rootCmd) {
		slog.Debug("registering MCP tool", "tool_name", tool.Name)
		mcp.AddTool(srv, tool, bridge.Execute)
	}

	if c.Transport == nil {
		c.Transport = &mcp.LoggingTransport{
			Transport: &mcp.StdioTransport{},
			Writer:    os.Stderr,
		}
	}

	return srv.Run(cmd.Context(), c.Transport)
}

// FromRootCmd converts a Cobra command tree into MCP tools.
func (c *Config) tools(rootCmd *cobra.Command) []*mcp.Tool {
	slog.Debug("starting tool generation from root command", "root_cmd", rootCmd.Name())
	tools := c.addTools(rootCmd, nil)
	slog.Info("tool generation completed", "total_tools", len(tools))
	return tools
}

func (c *Config) addTools(cmd *cobra.Command, tools []*mcp.Tool) []*mcp.Tool {
	if cmd == nil {
		return tools
	}

	// Register all subcommands
	for _, subCmd := range cmd.Commands() {
		tools = c.addTools(subCmd, tools)
	}

	// Apply all filters
	if c.Filters == nil {
		c.Filters = DefaultFilters()
	}
	for _, filter := range c.Filters {
		if !filter(cmd) {
			return tools
		}
	}

	tool := bridge.CreateToolFromCmd(cmd, c.ForOptions)

	slog.Debug("created tool", "tool_name", tool.Name, "description", tool.Description)
	return append(tools, tool)
}
