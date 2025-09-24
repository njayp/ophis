package ophis

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/njayp/ophis/internal/bridge"
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
	// If nil or empty, defaults to exposing all safe commands with all safe flags.
	Selectors []Selector

	// PreRun is middleware hook that runs before each tool call
	// Return a cancelled context to prevent execution.
	// Common uses: add timeouts, rate limiting, auth checks, metrics.
	PreRun func(context.Context, *mcp.CallToolRequest, bridge.CmdToolInput) (context.Context, *mcp.CallToolRequest, bridge.CmdToolInput)

	// PostRun is middleware hook that runs after each tool call
	// Common uses: error handling, response filtering, metrics collection.
	PostRun func(context.Context, *mcp.CallToolRequest, bridge.CmdToolInput, *mcp.CallToolResult, bridge.CmdToolOutput, error) (*mcp.CallToolResult, bridge.CmdToolOutput, error)

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

	if c.Selectors == nil {
		c.Selectors = []Selector{{}}
	}

	// get tools recursively
	return c.toolsRecursive(rootCmd, nil)
}

func (c *Config) toolsRecursive(cmd *cobra.Command, tools []*mcp.Tool) []*mcp.Tool {
	if cmd == nil {
		return tools
	}

	// register all subcommands
	for _, subCmd := range cmd.Commands() {
		tools = c.toolsRecursive(subCmd, tools)
	}

	// cycle through selectors until one matches the cmd
	for _, s := range c.Selectors {
		if s.cmdSelect(cmd) {
			// create tool with selected flags
			tool := bridge.CreateToolFromCmd(cmd, bridge.FlagSelector(s.FlagSelector))
			slog.Debug("created tool", "tool_name", tool.Name)
			return append(tools, tool)
		}
	}

	// no selectors matched
	return tools
}

func (c *Config) execute(ctx context.Context, request *mcp.CallToolRequest, input bridge.CmdToolInput) (result *mcp.CallToolResult, output bridge.CmdToolOutput, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()

	if c.PreRun != nil {
		ctx, request, input = c.PreRun(ctx, request, input)
	}

	result, output, err = bridge.Execute(ctx, request, input)

	if c.PostRun != nil {
		result, output, err = c.PostRun(ctx, request, input, result, output, err)
	}

	return
}
