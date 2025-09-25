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
	// If nil or empty, defaults to exposing all commands with all flags.
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

func (c *Config) bridgeSelectors() bridge.Selectors {
	// if selectors is empty or nil, return default selector
	length := len(c.Selectors)
	if length == 0 {
		return bridge.Selectors{{}}
	}

	selectors := make([]bridge.Selector, length)
	for i, s := range c.Selectors {
		selectors[i] = bridge.Selector{
			CmdSelector:  bridge.CmdSelector(s.CmdSelector),
			FlagSelector: bridge.FlagSelector(s.FlagSelector),
		}
	}

	return selectors
}

// tools takes care of setup, and calls toolsRecursive
func (c *Config) tools(rootCmd *cobra.Command) []*mcp.Tool {
	// slog to stderr
	handler := slog.NewTextHandler(os.Stderr, c.SloggerOptions)
	slog.SetDefault(slog.New(handler))

	return c.bridgeSelectors().ToolsRecursive(rootCmd, nil)
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
