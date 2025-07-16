package tools

import (
	"slices"

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

// Filter is a function type used by the Generator to filter commands.
// It returns true if the command should be included in the generated tools.
type Filter func(*cobra.Command) bool

// WithFilter adds a custom filter function to the generator.
func WithFilter(filter Filter) GeneratorOption {
	return func(g *Generator) {
		g.filters = append(g.filters, filter)
	}
}

// WithExclusions adds a filter to exclude listed command names from the generated tools.
func WithExclusions(list []string) GeneratorOption {
	return WithFilter(func(cmd *cobra.Command) bool {
		return !slices.Contains(list, cmd.Name())
	})
}

func withoutHidden() GeneratorOption {
	return WithFilter(func(cmd *cobra.Command) bool {
		return !cmd.Hidden
	})
}

// NewGenerator creates a new Generator with the specified options.
func NewGenerator(opts ...GeneratorOption) *Generator {
	g := &Generator{}

	WithExclusions([]string{MCPCommandName, "help", "completion"})(g)
	withoutHidden()(g)

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
	// Create the tool name
	toolName := cmd.Name()
	if parentPath != "" {
		toolName = parentPath + "_" + cmd.Name()
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
