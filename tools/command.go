// Package tools provides functionality for converting Cobra commands into MCP tools.
// It handles the registration and metadata generation for command-to-tool conversion.
package tools

import (
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func newTool(cmd *cobra.Command, toolName string) Tool {
	toolOptions := []mcp.ToolOption{
		mcp.WithDescription(descFromCmd(cmd)),
	}

	// add flags to tool
	flagMap := flagMapFromCmd(cmd)
	toolOptions = append(toolOptions, mcp.WithObject(FlagsParam,
		mcp.Description("flag options"),
		mcp.Properties(flagMap),
	))

	// Add an "args" parameter for positional arguments
	argsDescription := argsDescFromCmd(cmd)
	toolOptions = append(toolOptions, mcp.WithString(PositionalArgsParam,
		mcp.Description(argsDescription),
	))

	// Create the tool
	return Tool{
		Tool: mcp.NewTool(toolName, toolOptions...),
	}
}

func argsDescFromCmd(cmd *cobra.Command) string {
	argsDescription := "Positional arguments"
	if cmd.Use != "" {
		argsDescription += fmt.Sprintf(". Usage: %s", cmd.Use)
	}

	return argsDescription
}

func flagMapFromCmd(cmd *cobra.Command) map[string]any {
	// map for tool object
	flagMap := map[string]any{}

	// add local flags to flag map
	cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
		if flag.Hidden {
			return
		}

		flagMap[flag.Name] = flagToolOption(flag)
	})

	// add inherited flags to flag map
	cmd.InheritedFlags().VisitAll(func(flag *pflag.Flag) {
		if flag.Hidden {
			return
		}

		// Check if this flag was already added from local flags to avoid duplicates
		if _, ok := flagMap[flag.Name]; !ok {
			flagMap[flag.Name] = flagToolOption(flag)
		}
	})

	return flagMap
}

// descFromCmd creates a description for the MCP tool from the Cobra command
func descFromCmd(cmd *cobra.Command) string {
	desc := cmd.Long
	if desc == "" {
		desc = cmd.Short
	}

	return desc
}

func flagToolOption(flag *pflag.Flag) map[string]string {
	description := flag.Usage
	if description == "" {
		description = fmt.Sprintf("Flag: %s", flag.Name)
	}

	// Improve type detection for better MCP tool parameter definitions
	flagType := flag.Value.Type()
	switch flagType {
	case "stringSlice", "stringArray":
		return map[string]string{
			"type":        "stringArray",
			"description": description,
		}
	case "intSlice":
		return map[string]string{
			"type":        "intArray",
			"description": description,
		}
	case "bool":
		return map[string]string{
			"type":        "boolean",
			"description": description,
		}
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
		return map[string]string{
			"type":        "integer",
			"description": description,
		}
	case "float32", "float64":
		return map[string]string{
			"type":        "number",
			"description": description,
		}
	default:
		return map[string]string{
			"type":        "string",
			"description": description,
		}
	}
}
