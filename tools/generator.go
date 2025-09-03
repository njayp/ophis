package tools

import (
	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/spf13/cobra"
)

// Generator converts Cobra commands into MCP tools with configurable exclusions.
type Generator struct {
	filters []Filter
	handler Handler
}

// GeneratorOption is a function type for configuring Generator instances.
type GeneratorOption func(*Generator)

// NewGenerator creates a new Generator with the specified options.
//
// By default, the Generator:
//   - Excludes hidden commands
//   - Excludes "mcp", "help", and "completion" commands
//   - Uses DefaultHandler() which returns command output as plain text
//
// Available options:
//
//	WithFilters(filters ...Filter) - Replace all filters with custom ones
//	  Example: NewGenerator(WithFilters(Allow([]string{"get", "list"})))
//
//	AddFilter(filter Filter) - Add an additional filter to the existing ones
//	  Example: NewGenerator(AddFilter(Exclude([]string{"dangerous-cmd"})))
//
//	WithHandler(handler Handler) - Set a custom handler for processing command output
//	  Example: NewGenerator(WithHandler(myCustomHandler))
//
// Common filter functions:
//
//	Hidden() - Excludes hidden commands (applied by default)
//	Exclude([]string) - Excludes commands by name
//	Allow([]string) - Only includes commands whose path contains these names
func NewGenerator(opts ...GeneratorOption) *Generator {
	g := &Generator{
		// default filters
		filters: []Filter{
			Hidden(),
			Exclude([]string{MCPCommandName, "help", "completion"}),
		},
	}

	for _, opt := range opts {
		opt(g)
	}

	return g
}

// FromRootCmd recursively converts a Cobra command tree into MCP tools.
func (g *Generator) FromRootCmd(cmd *cobra.Command) []Controller {
	slog.Debug("starting tool generation from root command", "root_cmd", cmd.Name())
	tools := g.fromCmd(cmd, "", []Controller{})
	slog.Info("tool generation completed", "total_tools", len(tools))
	return tools
}

func (g *Generator) fromCmd(cmd *cobra.Command, parentPath string, tools []Controller) []Controller {
	if cmd == nil {
		return tools
	}

	// Create the tool name
	toolName := cmd.Name()
	if parentPath != "" {
		toolName = parentPath + "_" + toolName
	}

	slog.Debug("processing command", "command", toolName, "has_run", cmd.Run != nil || cmd.RunE != nil)

	// Register subcommands
outer:
	for _, subCmd := range cmd.Commands() {
		for _, filter := range g.filters {
			if !filter(subCmd) {
				// logging should be handled by the filter itself
				continue outer
			}
		}

		tools = g.fromCmd(subCmd, toolName, tools)
	}

	// Skip if the command has no runnable function
	if cmd.Run == nil && cmd.RunE == nil {
		slog.Debug("skipping command without run function", "command", toolName)
		return tools
	}

	toolOptions := toolOptsFromCmd(cmd)
	tool := Controller{
		Tool:    mcp.NewTool(toolName, toolOptions...),
		handler: g.handler, // Use the configured handler
	}

	slog.Debug("created tool", "tool_name", toolName, "description", tool.Tool.Description)
	return append(tools, tool)
}
