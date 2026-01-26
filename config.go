package ophis

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

// Config customizes MCP server behavior and command-to-tool conversion.
type Config struct {
	// Selectors defines rules for converting commands to MCP tools.
	// Each selector specifies which commands to match and which flags to include.
	//
	// Basic safety filters are always applied first:
	//   - Hidden/deprecated commands and flags are excluded
	//   - Non-runnable commands are excluded
	//   - Built-in commands (mcp, help, completion) are excluded
	//
	// Then selectors are evaluated in order for each command:
	//   1. The first selector whose CmdSelector returns true is used
	//   2. That selector's FlagSelector determines which flags are included
	//   3. If no selectors match, the command is not exposed as a tool
	//
	// If nil or empty, defaults to exposing all commands with all flags.
	Selectors []Selector

	// ToolNamePrefix replaces the root command name in tool names.
	// This is useful for shortening tool names to comply with API limits (e.g., Claude's 64 char limit).
	// For example, if root command is "omnistrate-ctl" and ToolNamePrefix is "omctl",
	// a command "omnistrate-ctl cost by-cell list" becomes "omctl_cost_by-cell_list" instead of
	// "omnistrate-ctl_cost_by-cell_list".
	// If empty, the root command name is used as-is.
	ToolNamePrefix string

	// SloggerOptions configures logging to stderr.
	// Default: Info level logging.
	SloggerOptions *slog.HandlerOptions

	// ServerOptions for the underlying MCP server.
	ServerOptions *mcp.ServerOptions

	// Transport for stdio transport configuration.
	Transport mcp.Transport

	server         *mcp.Server
	tools          []*mcp.Tool
	toolNamePrefix string // resolved prefix (either ToolNamePrefix or root command name)
}

func (c *Config) serveStdio(cmd *cobra.Command) error {
	if c.Transport == nil {
		c.Transport = &mcp.StdioTransport{}
	}

	c.registerTools(cmd)
	return c.server.Run(cmd.Context(), c.Transport)
}

func (c *Config) serveHTTP(cmd *cobra.Command, addr string) error {
	c.registerTools(cmd)

	// Create the streamable HTTP handler.
	handler := mcp.NewStreamableHTTPHandler(func(_ *http.Request) *mcp.Server {
		return c.server
	}, nil)

	server := &http.Server{Addr: addr, Handler: handler}

	// Shutdown gracefully
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		select {
		case <-ch:
		case <-cmd.Context().Done():
		}
		signal.Stop(ch)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			slog.Error("error shutting down server", "error", err)
		}
	}()

	cmd.Printf("MCP server listening on address %q\n", addr)
	return server.ListenAndServe()
}

// registerTools fully initializes a MCP server and populates c.tools
func (c *Config) registerTools(cmd *cobra.Command) {
	// slog to stderr
	handler := slog.NewTextHandler(os.Stderr, c.SloggerOptions)
	slog.SetDefault(slog.New(handler))

	// get root cmd
	rootCmd := cmd
	for rootCmd.Parent() != nil {
		rootCmd = rootCmd.Parent()
	}

	// resolve tool name prefix
	if c.ToolNamePrefix != "" {
		c.toolNamePrefix = c.ToolNamePrefix
	} else {
		c.toolNamePrefix = rootCmd.Name()
	}

	// make server
	c.server = mcp.NewServer(&mcp.Implementation{
		Name:    rootCmd.Name(),
		Version: rootCmd.Version,
	}, c.ServerOptions)

	// ensure at least one selector exists for tool creation logic
	if len(c.Selectors) == 0 {
		c.Selectors = []Selector{{}}
	}

	// register tools
	c.registerToolsRecursive(rootCmd)
}

// registerTools explores a cmd tree, making tools recursively out of the provided cmd and its children
func (c *Config) registerToolsRecursive(cmd *cobra.Command) {
	// register all subcommands
	for _, subCmd := range cmd.Commands() {
		c.registerToolsRecursive(subCmd)
	}

	// apply basic filters
	if cmdFilter(cmd) {
		return
	}

	// cycle through selectors until one matches the cmd
	for i, s := range c.Selectors {
		if s.CmdSelector != nil && !s.CmdSelector(cmd) {
			continue
		}

		// create tool from cmd
		tool := s.createToolFromCmd(cmd, c.toolNamePrefix)
		slog.Debug("created tool", "tool_name", tool.Name, "selector_index", i)

		// register tool with server
		mcp.AddTool(c.server, tool, s.execute)

		// add tool to manager's tool list (for `tools` command)
		c.tools = append(c.tools, tool)

		// only the first matching selector is used
		break
	}
}

// cmdFilter returns true if cmd should be filtered out
func cmdFilter(cmd *cobra.Command) bool {
	if cmd.Hidden || cmd.Deprecated != "" {
		return true
	}

	if cmd.Run == nil && cmd.RunE == nil && cmd.PreRun == nil && cmd.PreRunE == nil {
		return true
	}

	return AllowCmdsContaining("mcp", "help", "completion")(cmd)
}
