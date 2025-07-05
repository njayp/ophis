package tools

import (
	"slices"

	"github.com/spf13/cobra"
)

// FromRootCmd provides backward compatibility by creating a default generator
// and converting a Cobra command tree into MCP tools.
func FromRootCmd(cmd *cobra.Command) []Tool {
	generator := NewGenerator()
	return generator.FromRootCmd(cmd)
}

// Generator converts Cobra commands into MCP tools with configurable exclusions.
type Generator struct {
	exclusions []string
}

// GeneratorOption is a function type for configuring Generator instances.
type GeneratorOption func(*Generator)

// WithExclusions sets the list of command names to exclude from the generated tools.
func WithExclusions(exclusions []string) GeneratorOption {
	return func(g *Generator) {
		g.exclusions = append(g.exclusions, exclusions...)
	}
}

// NewGenerator creates a new Generator with the specified options.
func NewGenerator(opts ...GeneratorOption) *Generator {
	g := &Generator{
		exclusions: []string{MCPCommandName, "help"},
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
	// Create the tool name
	toolName := cmd.Name()
	if parentPath != "" {
		toolName = parentPath + "_" + cmd.Name()
	}

	// Register subcommands
	for _, subCmd := range cmd.Commands() {
		if subCmd.Hidden {
			continue
		}

		// ignore excluded commands
		if slices.Contains(g.exclusions, subCmd.Name()) {
			continue
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
