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

// CreateToolFromCmd creates an MCP tool from a Cobra command.
func CreateToolFromCmd(cmd *cobra.Command) *mcp.Tool {
	schema := inputSchema.copy()
	enhanceInputSchema(schema, cmd)

	// Create the tool
	return &mcp.Tool{
		Name:         generateToolName(cmd),
		Description:  buildToolDescription(cmd),
		InputSchema:  schema,
		OutputSchema: outputSchema.copy(),
	}
}

// generateToolName creates a tool name from the command path.
func generateToolName(cmd *cobra.Command) string {
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

// buildToolDescription creates a comprehensive tool description.
func buildToolDescription(cmd *cobra.Command) string {
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

// enhanceInputSchema modifies the input schema with command-specific flag information.
func enhanceInputSchema(schema *jsonschema.Schema, cmd *cobra.Command) {
	if schema.Properties == nil {
		return
	}

	// Enhance flags property
	if flagsSchema, exists := schema.Properties["flags"]; exists {
		enhanceFlagsSchema(flagsSchema, cmd)
	}

	// Enhance args property
	if argsSchema, exists := schema.Properties["args"]; exists {
		enhanceArgsSchema(argsSchema, cmd)
	}
}

// enhanceFlagsSchema adds detailed flag information to the flags property.
func enhanceFlagsSchema(schema *jsonschema.Schema, cmd *cobra.Command) {
	// Ensure properties map exists
	if schema.Properties == nil {
		schema.Properties = make(map[string]*jsonschema.Schema)
	}

	// Process local flags
	cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
		if flag.Hidden {
			return
		}
		addFlagToSchema(schema, flag)
	})

	// Process inherited flags
	cmd.InheritedFlags().VisitAll(func(flag *pflag.Flag) {
		if flag.Hidden {
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
	switch flag.Value.Type() {
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
		slog.Warn("unknown flag type, defaulting to string", "flag", flag.Name, "type", flag.Value.Type())
	}

	schema.Properties[flag.Name] = flagSchema
}

// enhanceArgsSchema adds detailed argument information to the args property.
func enhanceArgsSchema(schema *jsonschema.Schema, cmd *cobra.Command) {
	description := "Positional command line arguments"
	use := cmd.Use

	// remove "[flags]" from usage
	use = strings.Replace(use, " [flags]", "", 1)

	// Extract argument pattern from cmd.Use
	if use != "" {
		if spaceIdx := strings.IndexByte(use, ' '); spaceIdx != -1 {
			argsPattern := use[spaceIdx+1:]
			if argsPattern != "" {
				description += fmt.Sprintf("\n\nUsage pattern: %s", argsPattern)
			}
		}
	}

	schema.Description = description
}
