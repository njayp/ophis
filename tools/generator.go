package tools

import (
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
	}

	for _, opt := range opts {
		opt(g)
	}

	return g
}

// FromRootCmd recursively converts a Cobra command tree into MCP tools.
func (g *Generator) FromRootCmd(cmd *cobra.Command) []Tool {
	return g.fromCmd(cmd, "", []Tool{})
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

	// Register subcommands
outer:
	for _, subCmd := range cmd.Commands() {
		for _, filter := range g.filters {
			if !filter(subCmd) {
				continue outer
			}
		}

		tools = g.fromCmd(subCmd, toolName, tools)
	}

	// Skip if the command has no runnable function
	if cmd.Run == nil && cmd.RunE == nil {
		return tools
	}

	tools = append(tools, newTool(cmd, toolName))
	return tools
}
