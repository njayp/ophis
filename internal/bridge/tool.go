package bridge

import (
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Selector is a filter for flags
// Return true to include flag
type Selector func(*pflag.Flag) bool

// CreateToolFromCmd creates an MCP tool from a Cobra command.
func CreateToolFromCmd(cmd *cobra.Command, selector Selector) *mcp.Tool {
	schema := inputSchema.copy()
	enhanceFlagsSchema(schema.Properties["flags"], cmd, selector)
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
	// Count depth for capacity hint
	var names []string
	current := cmd
	for current != nil && current.Name() != "" {
		names = append(names, current.Name())
		current = current.Parent()
	}

	slices.Reverse(names)
	return strings.Join(names, "_")
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
func enhanceFlagsSchema(schema *jsonschema.Schema, cmd *cobra.Command, selector Selector) {
	// Ensure properties map exists
	if schema.Properties == nil {
		schema.Properties = make(map[string]*jsonschema.Schema)
	}

	// Process local flags
	cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
		if !selector(flag) {
			return
		}

		addFlagToSchema(schema, flag)
	})

	// Process inherited flags
	cmd.InheritedFlags().VisitAll(func(flag *pflag.Flag) {
		if !selector(flag) {
			return
		}

		// Skip if already added as local flag
		if _, exists := schema.Properties[flag.Name]; !exists {
			addFlagToSchema(schema, flag)
		}
	})
}

// addFlagToSchema adds a single flag to the schema properties.
func addFlagToSchema(schema *jsonschema.Schema, flag *pflag.Flag) {
	flagSchema := &jsonschema.Schema{
		Description: flag.Usage,
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
