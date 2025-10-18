package bridge

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

// Manager is an array of Selector, it is required to call ToolsRecursive
type Manager struct {
	Selectors []Selector
	Server    *mcp.Server
	Tools     []*mcp.Tool
	// Registered prompts, resources, and templates are tracked for export commands
	Prompts           []*mcp.Prompt
	Resources         []*mcp.Resource
	ResourceTemplates []*mcp.ResourceTemplate
}

// RegisterTools explores a cmd tree, making tools recursively out of the provided cmd and its children
func (m *Manager) RegisterTools(cmd *cobra.Command) {
	// register all subcommands
	for _, subCmd := range cmd.Commands() {
		m.RegisterTools(subCmd)
	}

	// cycle through selectors until one matches the cmd
	for i, s := range m.Selectors {
		if s.cmdSelect(cmd) {
			// create tool from cmd
			tool := s.CreateToolFromCmd(cmd)
			slog.Debug("created tool", "tool_name", tool.Name, "selector_index", i)

			// register tool with server
			mcp.AddTool(m.Server, tool, s.execute)

			// add tool to manager's tool list (for `tools` command)
			m.Tools = append(m.Tools, tool)

			// only the first matching selector is used
			break
		}
	}
}

// CreateToolFromCmd creates an MCP tool from a Cobra command.
func (s Selector) CreateToolFromCmd(cmd *cobra.Command) *mcp.Tool {
	schema := inputSchema.copy()
	s.enhanceFlagsSchema(schema.Properties["flags"], cmd)
	enhanceArgsSchema(schema.Properties["args"], cmd)

	// Create the tool
	return &mcp.Tool{
		Name:         toolName(cmd),
		Description:  toolDescription(cmd),
		InputSchema:  schema,
		OutputSchema: outputSchema.copy(),
	}
}

// toolName creates a tool name from the command path.
func toolName(cmd *cobra.Command) string {
	path := cmd.CommandPath()
	return strings.ReplaceAll(path, " ", "_")
}

// toolDescription creates a comprehensive tool description.
func toolDescription(cmd *cobra.Command) string {
	var parts []string

	// Use Long description if available, otherwise Short
	if cmd.Long != "" {
		parts = append(parts, cmd.Long)
	} else if cmd.Short != "" {
		parts = append(parts, cmd.Short)
	} else {
		parts = append(parts, fmt.Sprintf("Execute the %s command", cmd.Name()))
	}

	// Add examples if available
	if cmd.Example != "" {
		parts = append(parts, fmt.Sprintf("Examples:\n%s", cmd.Example))
	}

	return strings.Join(parts, "\n")
}

// enhanceArgsSchema adds detailed argument information to the args property.
func enhanceArgsSchema(schema *jsonschema.Schema, cmd *cobra.Command) {
	description := "Positional command line arguments"

	// remove "[flags]" from usage
	usage := strings.Replace(cmd.Use, " [flags]", "", 1)

	// Extract argument pattern from cmd.Use
	if usage != "" {
		if spaceIdx := strings.IndexByte(usage, ' '); spaceIdx != -1 {
			argsPattern := usage[spaceIdx+1:]
			if argsPattern != "" {
				description += fmt.Sprintf("\nUsage pattern: %s", argsPattern)
			}
		}
	}

	schema.Description = description
}
