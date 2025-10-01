package bridge

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Manager is an array of Selector, it is required to call ToolsRecursive
type Manager struct {
	Selectors []Selector
	Server    *mcp.Server
	Tools     []*mcp.Tool
}

// RegisterTools explores a cmd tree, making tools recursively out of the provided cmd and its children
func (m *Manager) RegisterTools(cmd *cobra.Command) {
	if cmd == nil {
		slog.Error("ToolsRecursive called with nil command")
		return
	}

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

// enhanceFlagsSchema adds detailed flag information to the flags property.
func (s Selector) enhanceFlagsSchema(schema *jsonschema.Schema, cmd *cobra.Command) {
	// Ensure properties map exists
	if schema.Properties == nil {
		schema.Properties = make(map[string]*jsonschema.Schema)
	}

	// Process local flags
	cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
		if s.localFlagSelect(flag) {
			addFlagToSchema(schema, flag)
		}
	})

	// Process inherited flags
	cmd.InheritedFlags().VisitAll(func(flag *pflag.Flag) {
		if s.inheritedFlagSelect(flag) {
			// Skip if already added as local flag
			if _, exists := schema.Properties[flag.Name]; !exists {
				addFlagToSchema(schema, flag)
			}
		}
	})
}

// isFlagRequired checks if a flag has been marked as required by Cobra.
// Cobra uses the BashCompOneRequiredFlag annotation to track required flags.
func isFlagRequired(flag *pflag.Flag) bool {
	if flag.Annotations == nil {
		return false
	}

	// Check if the flag has the required annotation
	// The constant is defined as "cobra_annotation_bash_completion_one_required_flag" in Cobra
	if val, ok := flag.Annotations[cobra.BashCompOneRequiredFlag]; ok {
		// The annotation is present if the flag is required
		return len(val) > 0 && val[0] == "true"
	}

	return false
}

// addFlagToSchema adds a single flag to the schema properties.
func addFlagToSchema(schema *jsonschema.Schema, flag *pflag.Flag) {
	flagSchema := &jsonschema.Schema{
		Description: flag.Usage,
	}

	// Check if flag is marked as required in its annotations
	// Cobra uses the BashCompOneRequiredFlag annotation to mark required flags
	if isRequired := isFlagRequired(flag); isRequired {
		// Mark the flag as required in the schema
		if schema.Required == nil {
			schema.Required = []string{}
		}

		schema.Required = append(schema.Required, flag.Name)
	}

	// Set appropriate JSON schema type based on flag type
	t := flag.Value.Type()
	switch t {
	case "bool":
		flagSchema.Type = "boolean"
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "count":
		flagSchema.Type = "integer"
	case "float32", "float64":
		flagSchema.Type = "number"
	case "string":
		flagSchema.Type = "string"
	case "stringSlice", "stringArray":
		flagSchema.Type = "array"
		flagSchema.Items = &jsonschema.Schema{Type: "string"}
	case "intSlice", "int32Slice", "int64Slice", "uintSlice":
		flagSchema.Type = "array"
		flagSchema.Items = &jsonschema.Schema{Type: "integer"}
	case "float32Slice", "float64Slice":
		flagSchema.Type = "array"
		flagSchema.Items = &jsonschema.Schema{Type: "number"}
	case "boolSlice":
		flagSchema.Type = "array"
		flagSchema.Items = &jsonschema.Schema{Type: "boolean"}
	case "duration":
		// Duration is represented as a string in Go's duration format
		flagSchema.Type = "string"
		flagSchema.Description += " (format: Go duration string, e.g., '10s', '2h45m')"
		flagSchema.Pattern = `^-?([0-9]+(\.[0-9]+)?(ns|us|µs|ms|s|m|h))+$`
	case "ip":
		flagSchema.Type = "string"
		flagSchema.Description += " (format: IPv4 or IPv6 address)"
		flagSchema.Pattern = `^((25[0-5]|(2[0-4]|1\d|[1-9]|)\d)\.){3}(25[0-5]|(2[0-4]|1\d|[1-9]|)\d)$|^(([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4})$`
	case "ipMask":
		flagSchema.Type = "string"
		flagSchema.Description += " (format: IP mask, e.g., '255.255.255.0')"
	case "ipNet":
		flagSchema.Type = "string"
		flagSchema.Description += " (format: CIDR notation, e.g., '192.168.1.0/24')"
		flagSchema.Pattern = `^((25[0-5]|(2[0-4]|1\d|[1-9]|)\d)\.){3}(25[0-5]|(2[0-4]|1\d|[1-9]|)\d)/([0-9]|[1-2][0-9]|3[0-2])$`
	case "bytesHex":
		flagSchema.Type = "string"
		flagSchema.Description += " (format: hexadecimal string)"
		flagSchema.Pattern = `^[0-9a-fA-F]*$`
	case "bytesBase64":
		flagSchema.Type = "string"
		flagSchema.Description += " (format: base64 encoded string)"
		flagSchema.Pattern = `^[A-Za-z0-9+/]*={0,2}$`
	default:
		// Default to string for unknown types
		flagSchema.Type = "string"
		flagSchema.Description += fmt.Sprintf(" (type: %s)", t)
		slog.Debug("unknown flag type, defaulting to string", "flag", flag.Name, "type", t)
	}

	schema.Properties[flag.Name] = flagSchema
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
