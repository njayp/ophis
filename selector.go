package ophis

import (
	"context"
	"fmt"

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

// PreRunFunc is middleware hook that runs before each tool call
// Return a cancelled context to prevent execution.
// Common uses: add timeouts, rate limiting, auth checks, metrics.
type PreRunFunc func(context.Context, *mcp.CallToolRequest, ToolInput) (context.Context, *mcp.CallToolRequest, ToolInput)

// PostRunFunc is middleware hook that runs after each tool call
// Common uses: error handling, response filtering, metrics collection.
type PostRunFunc func(context.Context, *mcp.CallToolRequest, ToolInput, *mcp.CallToolResult, ToolOutput, error) (*mcp.CallToolResult, ToolOutput, error)

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

	// PreRun is middleware hook that runs before each tool call
	// Return a cancelled context to prevent execution.
	// Common uses: add timeouts, rate limiting, auth checks, metrics.
	PreRun PreRunFunc

	// PostRun is middleware hook that runs after each tool call
	// Common uses: error handling, response filtering, metrics collection.
	PostRun PostRunFunc
}

// cmdSelect returns true if the command passes default filters and this selector's CmdSelector (if any).
func (s *Selector) cmdSelect(cmd *cobra.Command) bool {
	if cmd.Hidden || cmd.Deprecated != "" {
		return false
	}

	if cmd.Run == nil && cmd.RunE == nil && cmd.PreRun == nil && cmd.PreRunE == nil {
		return false
	}

	if AllowCmdsContaining("mcp", "help", "completion")(cmd) {
		return false
	}

	if s.CmdSelector != nil {
		return s.CmdSelector(cmd)
	}

	return true
}

func defaultFlagSelect(flag *pflag.Flag) bool {
	return !flag.Hidden && flag.Deprecated == ""
}

// localFlagSelect returns true if the flag passes default filters and this selector's FlagSelector (if any).
func (s *Selector) localFlagSelect(flag *pflag.Flag) bool {
	if !defaultFlagSelect(flag) {
		return false
	}

	if s.LocalFlagSelector != nil {
		return s.LocalFlagSelector(flag)
	}

	return true
}

// flagSelect returns true if the flag passes default filters and this selector's FlagSelector (if any).
func (s *Selector) inheritedFlagSelect(flag *pflag.Flag) bool {
	if !defaultFlagSelect(flag) {
		return false
	}

	if s.InheritedFlagSelector != nil {
		return s.InheritedFlagSelector(flag)
	}

	return true
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
			flags.AddFlagToSchema(schema, flag)
		}
	})

	// Process inherited flags
	cmd.InheritedFlags().VisitAll(func(flag *pflag.Flag) {
		if s.inheritedFlagSelect(flag) {
			// Skip if already added as local flag
			if _, exists := schema.Properties[flag.Name]; !exists {
				flags.AddFlagToSchema(schema, flag)
			}
		}
	})

	// Set AdditionalProperties to false
	// See https://github.com/google/jsonschema-go/issues/13
	schema.AdditionalProperties = &jsonschema.Schema{Not: &jsonschema.Schema{}}
}

// createToolFromCmd creates an MCP tool from a Cobra command.
func (s Selector) createToolFromCmd(cmd *cobra.Command) *mcp.Tool {
	schema := inputSchema.Copy()
	s.enhanceFlagsSchema(schema.Properties["flags"], cmd)
	enhanceArgsSchema(schema.Properties["args"], cmd)

	// Create the tool
	return &mcp.Tool{
		Name:         toolName(cmd),
		Description:  toolDescription(cmd),
		InputSchema:  schema,
		OutputSchema: outputSchema.Copy(),
	}
}

func (s *Selector) execute(ctx context.Context, request *mcp.CallToolRequest, input ToolInput) (result *mcp.CallToolResult, output ToolOutput, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()

	if s.PreRun != nil {
		ctx, request, input = s.PreRun(ctx, request, input)
	}

	result, output, err = execute(ctx, request, input)

	if s.PostRun != nil {
		result, output, err = s.PostRun(ctx, request, input, result, output, err)
	}

	return result, output, err
}
