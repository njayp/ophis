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

// CreateToolFromCmd creates an MCP tool from a Cobra command.
func CreateToolFromCmd(cmd *cobra.Command, opts *jsonschema.ForOptions) *mcp.Tool {
	// Generate the tool name from command path
	toolName := generateToolName(cmd)

	// Generate base input schema
	inputSchema, err := jsonschema.For[CmdToolInput](opts)
	if err != nil {
		slog.Error("failed to generate input schema", "tool", toolName, "error", err)
		panic(fmt.Sprintf("Failed to generate input schema for %s: %v", toolName, err))
	}

	// Enhance the schema with command-specific information
	enhanceInputSchema(inputSchema, cmd)

	// Create the tool
	return &mcp.Tool{
		Name:        toolName,
		Description: buildToolDescription(cmd),
		InputSchema: inputSchema,
	}
}

// generateToolName creates a tool name from the command path.
func generateToolName(cmd *cobra.Command) string {
	var names []string
	current := cmd

	// Walk up the command tree to build the full path
	for current != nil && current.Name() != "" {
		names = append([]string{current.Name()}, names...)
		current = current.Parent()
	}

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

	// Add usage information
	if cmd.Use != "" {
		parts = append(parts, fmt.Sprintf("\nUsage: %s", cmd.Use))
	}

	// Add examples if available
	if cmd.Example != "" {
		parts = append(parts, fmt.Sprintf("\nExamples:\n%s", cmd.Example))
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
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
		flagSchema.Type = "integer"
	case "float32", "float64":
		flagSchema.Type = "number"
	case "stringSlice", "stringArray":
		flagSchema.Type = "array"
		flagSchema.Items = &jsonschema.Schema{Type: "string"}
	case "intSlice":
		flagSchema.Type = "array"
		flagSchema.Items = &jsonschema.Schema{Type: "integer"}
	default:
		flagSchema.Type = "string"
	}

	schema.Properties[flag.Name] = flagSchema
}

// enhanceArgsSchema adds detailed argument information to the args property.
func enhanceArgsSchema(schema *jsonschema.Schema, cmd *cobra.Command) {
	description := "Positional command line arguments"

	// Extract argument pattern from cmd.Use
	if cmd.Use != "" {
		if spaceIdx := strings.IndexByte(cmd.Use, ' '); spaceIdx != -1 {
			argsPattern := cmd.Use[spaceIdx+1:]
			if argsPattern != "" {
				description += fmt.Sprintf("\n\nUsage pattern: %s", argsPattern)
			}
		}
	}

	schema.Description = description
}
