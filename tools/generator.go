package tools

import (
	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/njayp/ophis/internal/cfgmgr"
	"github.com/spf13/cobra"
)

// Generator converts Cobra commands into MCP tools with configurable filtering and output handling.
type Generator struct {
	filters []Filter
	handler Handler
}

// GeneratorOption configures a Generator instance.
type GeneratorOption func(*Generator)

// NewGenerator creates a Generator with custom options.
//
// Default behavior:
//   - Excludes non-runnable commands (Runs filter)
//   - Excludes hidden commands (Hidden filter)
//   - Excludes "mcp", "help", and "completion" commands
//   - Returns command output as plain text (DefaultHandler)
func NewGenerator(opts ...GeneratorOption) *Generator {
	g := &Generator{
		// default filters
		filters: []Filter{
			Runs(),
			Hidden(),
			Exclude([]string{cfgmgr.MCPCommandName, "help", "completion"}),
		},
	}

	for _, opt := range opts {
		opt(g)
	}

	return g
}

// FromRootCmd converts a Cobra command tree into MCP tools.
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

	// Register all subcommands
	for _, subCmd := range cmd.Commands() {
		tools = g.fromCmd(subCmd, toolName, tools)
	}

	// Apply all filters
	for _, filter := range g.filters {
		if !filter(cmd) {
			return tools
		}
	}

	toolOptions := toolOptsFromCmd(cmd)
	tool := Controller{
		Tool:    mcp.NewTool(toolName, toolOptions...),
		handler: g.handler, // Use the configured handler
	}

	slog.Debug("created tool", "tool_name", toolName, "description", tool.Tool.Description)
	return append(tools, tool)
}
