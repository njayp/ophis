package tools

import (
	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/spf13/cobra"
)

// FromRootCmd creates a default generator and converts a Cobra command tree into MCP tools.
func FromRootCmd(cmd *cobra.Command) []Tool {
	generator := NewGenerator()
	return generator.FromRootCmd(cmd)
}

// Generator converts Cobra commands into MCP tools with configurable exclusions.
type Generator struct {
	filters []Filter
	handler Handler
}

// GeneratorOption is a function type for configuring Generator instances.
type GeneratorOption func(*Generator)

// NewGenerator creates a new Generator with the specified options.
func NewGenerator(opts ...GeneratorOption) *Generator {
	g := &Generator{
		// default filters
		filters: []Filter{
			Hidden(),
			Exclude([]string{MCPCommandName, "help", "completion"}),
		},
		handler: DefaultHandler(),
	}

	for _, opt := range opts {
		opt(g)
	}

	return g
}

// FromRootCmd recursively converts a Cobra command tree into MCP tools.
func (g *Generator) FromRootCmd(cmd *cobra.Command) []Tool {
	slog.Debug("starting tool generation from root command", "root_cmd", cmd.Name())
	tools := g.fromCmd(cmd, "", []Tool{})
	slog.Info("tool generation completed", "total_tools", len(tools))
	return tools
}

func (g *Generator) fromCmd(cmd *cobra.Command, parentPath string, tools []Tool) []Tool {
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

	toolOptions := newTool(cmd)
	tool := Tool{
		Tool:    mcp.NewTool(toolName, toolOptions...),
		Handler: g.handler, // Use the configured handler
	}

	slog.Debug("created tool", "tool_name", toolName, "description", tool.Tool.Description)
	return append(tools, tool)
}
