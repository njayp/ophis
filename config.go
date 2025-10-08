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
}

func (c *Config) serveStdio(cmd *cobra.Command) error {
	if c.Transport == nil {
		c.Transport = &mcp.StdioTransport{}
	}

	return c.manager(cmd).Server.Run(cmd.Context(), c.Transport)
}

func (c *Config) tools(cmd *cobra.Command) []*mcp.Tool {
	return c.manager(cmd).Tools
}

// manager fully initializes a bridge.Manager
func (c *Config) manager(cmd *cobra.Command) *bridge.Manager {
	// slog to stderr
	handler := slog.NewTextHandler(os.Stderr, c.SloggerOptions)
	slog.SetDefault(slog.New(handler))

	// get root cmd
	rootCmd := cmd
	for rootCmd.Parent() != nil {
		rootCmd = rootCmd.Parent()
	}

	// make server
	server := mcp.NewServer(&mcp.Implementation{
		Name:    rootCmd.Name(),
		Version: rootCmd.Version,
	}, c.ServerOptions)

	// make manager
	manager := &bridge.Manager{
		Selectors: c.selectors(),
		Server:    server,
	}

	// register tools
	manager.RegisterTools(rootCmd)
	return manager
}

// selectors converts Config.Selectors to bridge.Selectors
func (c *Config) selectors() []bridge.Selector {
	// if selectors is empty or nil, return default selector
	length := len(c.Selectors)
	if length == 0 {
		return []bridge.Selector{{}}
	}

	selectors := make([]bridge.Selector, length)
	for i, s := range c.Selectors {
		selectors[i] = bridge.Selector{
			CmdSelector:           bridge.CmdSelector(s.CmdSelector),
			LocalFlagSelector:     bridge.FlagSelector(s.LocalFlagSelector),
			InheritedFlagSelector: bridge.FlagSelector(s.InheritedFlagSelector),
			PreRun:                bridge.PreRunFunc(s.PreRun),
			PostRun:               bridge.PostRunFunc(s.PostRun),
		}
	}

	return selectors
}
