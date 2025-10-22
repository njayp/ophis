package ophis

import (
	"log/slog"
	"os"

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

	// SloggerOptions configures logging to stderr.
	// Default: Info level logging.
	SloggerOptions *slog.HandlerOptions

	// ServerOptions for the underlying MCP server.
	ServerOptions *mcp.ServerOptions

	// Transport for stdio transport configuration.
	Transport mcp.Transport

	server *mcp.Server
	tools  []*mcp.Tool
}

func (c *Config) serveStdio(cmd *cobra.Command) error {
	if c.Transport == nil {
		c.Transport = &mcp.StdioTransport{}
	}

	c.registerTools(cmd)
	return c.server.Run(cmd.Context(), c.Transport)
}

// registerTools fully initializes a MCP server
func (c *Config) registerTools(cmd *cobra.Command) {
	// slog to stderr
	handler := slog.NewTextHandler(os.Stderr, c.SloggerOptions)
	slog.SetDefault(slog.New(handler))

	// get root cmd
	rootCmd := cmd
	for rootCmd.Parent() != nil {
		rootCmd = rootCmd.Parent()
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

	// cycle through selectors until one matches the cmd
	for i, s := range c.Selectors {
		if s.cmdSelect(cmd) {
			// create tool from cmd
			tool := s.createToolFromCmd(cmd)
			slog.Debug("created tool", "tool_name", tool.Name, "selector_index", i)

			// register tool with server
			mcp.AddTool(c.server, tool, s.execute)

			// add tool to manager's tool list (for `tools` command)
			c.tools = append(c.tools, tool)

			// only the first matching selector is used
			break
		}
	}
}
