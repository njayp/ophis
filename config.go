package ophis

import (
	"log/slog"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/njayp/ophis/internal/bridge"
	"github.com/spf13/cobra"
)

// Prompt binds an MCP Prompt to its handler for registration.
type Prompt struct {
	Prompt  *mcp.Prompt
	Handler mcp.PromptHandler
}

// Resource binds an MCP Resource to its handler for registration.
type Resource struct {
	Resource *mcp.Resource
	Handler  mcp.ResourceHandler
}

// ResourceTemplate binds an MCP ResourceTemplate to its handler for registration.
type ResourceTemplate struct {
	Template *mcp.ResourceTemplate
	Handler  mcp.ResourceHandler
}

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

	// Prompts to register with the MCP server.
	Prompts []Prompt

	// Resources to register with the MCP server.
	Resources []Resource

	// ResourceTemplates to register with the MCP server.
	ResourceTemplates []ResourceTemplate
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

// prompts returns the registered prompts for the current server.
func (c *Config) prompts(cmd *cobra.Command) []*mcp.Prompt {
	return c.manager(cmd).Prompts
}

// resources returns the registered resources for the current server.
func (c *Config) resources(cmd *cobra.Command) []*mcp.Resource {
	return c.manager(cmd).Resources
}

// resourceTemplates returns the registered resource templates for the current server.
func (c *Config) resourceTemplates(cmd *cobra.Command) []*mcp.ResourceTemplate {
	return c.manager(cmd).ResourceTemplates
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

	// register prompts
	for _, p := range c.Prompts {
		if p.Prompt != nil && p.Handler != nil {
			server.AddPrompt(p.Prompt, p.Handler)
			manager.Prompts = append(manager.Prompts, p.Prompt)
		}
	}

	// register resources
	for _, r := range c.Resources {
		if r.Resource != nil && r.Handler != nil {
			server.AddResource(r.Resource, r.Handler)
			manager.Resources = append(manager.Resources, r.Resource)
		}
	}

	// register resource templates
	for _, rt := range c.ResourceTemplates {
		if rt.Template != nil && rt.Handler != nil {
			server.AddResourceTemplate(rt.Template, rt.Handler)
			manager.ResourceTemplates = append(manager.ResourceTemplates, rt.Template)
		}
	}
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
