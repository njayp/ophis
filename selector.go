package ophis

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/njayp/ophis/internal/bridge/flags"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// CmdSelector determines if a command should become an MCP tool.
// Return true to include the command as a tool.
// Note: Basic safety filters (hidden, deprecated, non-runnable) are always applied first.
// Commands are tested against selectors in order; the first matching selector wins.
type CmdSelector func(*cobra.Command) bool

// FlagSelector determines if a flag should be included in an MCP tool.
// Return true to include the flag.
// Note: Hidden and deprecated flags are always excluded regardless of this selector.
// This selector is only applied to commands that match the associated CmdSelector.
type FlagSelector func(*pflag.Flag) bool

// MiddlewareFunc is middleware hook that runs after each tool call
// Common uses: error handling, response filtering, metrics collection.
type MiddlewareFunc func(context.Context, *mcp.CallToolRequest, ToolInput, ExecuteFunc) (*mcp.CallToolResult, ToolOutput, error)

// ExecuteFunc defines the function signature for executing a tool.
type ExecuteFunc func(context.Context, *mcp.CallToolRequest, ToolInput) (*mcp.CallToolResult, ToolOutput, error)

// Selector contains selectors for filtering commands and flags.
// When multiple selectors are configured, they are evaluated in order.
// The first selector whose CmdSelector matches a command is used,
// and its FlagSelector determines which flags are included for that command.
//
// Basic safety filters are always applied automatically:
//   - Hidden/deprecated commands and flags are excluded
//   - Non-runnable commands are excluded
//   - Built-in commands (mcp, help, completion) are excluded
//
// This allows fine-grained control within safe boundaries, such as:
//   - Exposing different flags for different command groups
//   - Applying stricter flag filtering to dangerous commands
//   - Having a default catch-all selector with common flag exclusions
type Selector struct {
	// CmdSelector determines if this selector applies to a command.
	// If nil, accepts all commands that pass basic safety filters.
	// Cannot be used to bypass safety filters (hidden, deprecated, non-runnable).
	CmdSelector CmdSelector

	// LocalFlagSelector determines which flags to include for commands matched by CmdSelector.
	// If nil, includes all flags that pass basic safety filters.
	// Cannot be used to bypass safety filters (hidden, deprecated flags).
	LocalFlagSelector FlagSelector

	// InheritedFlagSelector determines which persistent flags to include for commands matched by CmdSelector.
	// If nil, includes all flags that pass basic safety filters.
	// Cannot be used to bypass safety filters (hidden, deprecated flags).
	InheritedFlagSelector FlagSelector

	// Middleware is an optional middleware hook that wraps around tool execution.
	// Common uses: error handling, response filtering, metrics collection.
	// If nil, no middleware is applied.
	Middleware MiddlewareFunc
}

// enhanceFlagsSchema adds detailed flag information to the flags property.
func (s Selector) enhanceFlagsSchema(schema *jsonschema.Schema, cmd *cobra.Command) {
	// Ensure properties map exists
	if schema.Properties == nil {
		schema.Properties = make(map[string]*jsonschema.Schema)
	}

	// basic filters
	filter := func(flag *pflag.Flag) bool {
		return flag.Hidden || flag.Deprecated != ""
	}

	// Process local flags
	cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
		if filter(flag) {
			return
		}

		if s.LocalFlagSelector != nil && !s.LocalFlagSelector(flag) {
			return
		}

		flags.AddFlagToSchema(schema, flag)
	})

	// Process inherited flags
	cmd.InheritedFlags().VisitAll(func(flag *pflag.Flag) {
		// Skip if already added as local flag
		if _, exists := schema.Properties[flag.Name]; exists {
			return
		}

		if filter(flag) {
			return
		}

		if s.InheritedFlagSelector != nil && !s.InheritedFlagSelector(flag) {
			return
		}

		flags.AddFlagToSchema(schema, flag)
	})

	// Set AdditionalProperties to false
	// See https://github.com/google/jsonschema-go/issues/13
	schema.AdditionalProperties = &jsonschema.Schema{Not: &jsonschema.Schema{}}
}

// createToolFromCmd creates an MCP tool from a Cobra command.
// The toolNamePrefix is used to replace the root command name in the tool name.
func (s Selector) createToolFromCmd(cmd *cobra.Command, toolNamePrefix string, inferArgConstraints bool) *mcp.Tool {
	schema := inputSchema.Copy()
	s.enhanceFlagsSchema(schema.Properties["flags"], cmd)
	enhanceArgsSchema(schema, cmd, inferArgConstraints)

	// Create the tool
	return &mcp.Tool{
		Name:         toolName(cmd, toolNamePrefix),
		Description:  toolDescription(cmd),
		InputSchema:  schema,
		OutputSchema: outputSchema.Copy(),
		Annotations:  toolAnnotations(cmd),
	}
}

// enhanceArgsSchema adds detailed argument information to the args property.
func enhanceArgsSchema(input *jsonschema.Schema, cmd *cobra.Command, inferConstraints bool) {
	schema := input.Properties["args"]
	description := "Positional command line arguments"

	// remove "[flags]" from usage
	usage := strings.ReplaceAll(cmd.Use, " [flags]", "")

	// Extract argument pattern from cmd.Use
	if usage != "" {
		if spaceIdx := strings.IndexByte(usage, ' '); spaceIdx != -1 {
			argsPattern := usage[spaceIdx+1:]
			if argsPattern != "" {
				description += fmt.Sprintf("\nUsage pattern: %s", argsPattern)
				if inferConstraints && cmd.Args != nil {
					minItems, maxItems, ok := argumentBounds(argsPattern)
					if !ok {
						schema.Description = description
						return
					}
					minConfirmed := false
					for count := 0; minItems > 0 && !minConfirmed && count <= minItems; count++ {
						accepted, probed := validatorAllowsArgCount(cmd, count)
						if !probed || accepted != (count == minItems) {
							break
						}
						minConfirmed = count == minItems
					}
					if minItems > 0 && minConfirmed {
						schema.MinItems = &minItems
					}
					if maxItems != nil {
						atMax, maxProbed := minConfirmed, minConfirmed
						if *maxItems != minItems || !minConfirmed {
							atMax, maxProbed = validatorAllowsArgCount(cmd, *maxItems)
						}
						if maxProbed {
							aboveMax, aboveProbed := validatorAllowsArgCount(cmd, *maxItems+1)
							if aboveProbed && atMax && !aboveMax {
								schema.MaxItems = maxItems
							}
						}
					}
					if minItems > 0 && minConfirmed {
						schema.Type = "array"
						schema.Types = nil
						input.Required = append(input.Required, "args")
					}
				}
			}
		}
	}

	schema.Description = description
}

// validatorAllowsArgCount invokes user code during registration with synthetic placeholder arguments.
func validatorAllowsArgCount(cmd *cobra.Command, count int) (accepted, probed bool) {
	probe := *cmd
	probe.SetOut(io.Discard)
	probe.SetErr(io.Discard)
	args := make([]string, count)
	for i := range args {
		args[i] = "arg"
	}
	defer func() {
		if r := recover(); r != nil {
			slog.Warn("args validator panicked during probe, skipping bounds", "command", cmd.CommandPath(), "count", count, "panic", r)
			probed = false
		}
	}()
	return cmd.Args(&probe, args) == nil, true
}

func argumentBounds(pattern string) (int, *int, bool) {
	fields := strings.Fields(pattern)
	minItems, maxItems := 0, 0
	variadic := false
	for i := 0; i < len(fields); i++ {
		if fields[i] == "..." {
			variadic = true
			continue
		}

		argument := fields[i]
		if strings.HasSuffix(argument, "...") {
			variadic = true
			argument = strings.TrimSuffix(argument, "...")
		}
		optional := strings.HasPrefix(argument, "[")
		closing := ""
		switch {
		case optional:
			closing = "]"
		case strings.HasPrefix(argument, "<"):
			closing = ">"
		case strings.ContainsAny(argument, "[]<>"):
			return 0, nil, false
		}
		for closing != "" && !strings.HasSuffix(argument, closing) {
			i++
			if i == len(fields) {
				return 0, nil, false
			}
			argument += " " + fields[i]
		}
		if strings.ContainsAny(argument, "()|") {
			return 0, nil, false
		}
		content := strings.Trim(argument, "[]<>")
		if content == "" || strings.ContainsAny(content, "[]<>") {
			return 0, nil, false
		}
		for _, field := range strings.Fields(content) {
			if strings.HasPrefix(field, "-") {
				return 0, nil, false
			}
		}
		if strings.Contains(argument, "...") {
			variadic = true
		}
		if !optional {
			minItems++
		}
		maxItems++
	}
	if variadic {
		return minItems, nil, true
	}
	return minItems, &maxItems, true
}

// toolName creates a tool name from the command path.
// The toolNamePrefix replaces the root command name in the path.
// For example, if the command path is "omnistrate-ctl cost by-cell list" and
// toolNamePrefix is "omctl", the result is "omctl_cost_by-cell_list".
func toolName(cmd *cobra.Command, toolNamePrefix string) string {
	path := cmd.CommandPath()

	// Replace the root command name with the prefix
	// The root command name is the first word in the path
	if spaceIdx := strings.IndexByte(path, ' '); spaceIdx != -1 {
		// Replace root command name with prefix, keep the rest
		path = toolNamePrefix + path[spaceIdx:]
	} else {
		// Single command (root command itself)
		path = toolNamePrefix
	}

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

func (s *Selector) execute(ctx context.Context, request *mcp.CallToolRequest, input ToolInput) (_ *mcp.CallToolResult, _ ToolOutput, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()

	if s.Middleware != nil {
		return s.Middleware(ctx, request, input, execute)
	}

	return execute(ctx, request, input)
}
