package bridge

import (
	"context"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/njayp/ophis/internal/cfgmgr"
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
type PreRunFunc func(context.Context, *mcp.CallToolRequest, CmdToolInput) (context.Context, *mcp.CallToolRequest, CmdToolInput)

// PostRunFunc is middleware hook that runs after each tool call
// Common uses: error handling, response filtering, metrics collection.
type PostRunFunc func(context.Context, *mcp.CallToolRequest, CmdToolInput, *mcp.CallToolResult, CmdToolOutput, error) (*mcp.CallToolResult, CmdToolOutput, error)

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
	// FlagSelector determines which flags to include for commands matched by CmdSelector.
	// If nil, includes all flags that pass basic safety filters.
	// Cannot be used to bypass safety filters (hidden, deprecated flags).
	FlagSelector FlagSelector

	// PreRun is middleware hook that runs before each tool call
	// Return a cancelled context to prevent execution.
	// Common uses: add timeouts, rate limiting, auth checks, metrics.
	PreRun PreRunFunc

	// PostRun is middleware hook that runs after each tool call
	// Common uses: error handling, response filtering, metrics collection.
	PostRun PostRunFunc
}

func defaultCmdSelect(c *cobra.Command) bool {
	if c.Hidden {
		return false
	}

	if c.Deprecated != "" {
		return false
	}

	if c.Run == nil && c.RunE == nil && c.PreRun == nil && c.PreRunE == nil {
		return false
	}

	return excludeCmd(cfgmgr.MCPCommandName, "help", "completion")(c)
}

// cmdSelect returns true if the command passes default filters and this selector's CmdSelector (if any).
func (s *Selector) cmdSelect(cmd *cobra.Command) bool {
	if !defaultCmdSelect(cmd) {
		return false
	}

	if s.CmdSelector != nil {
		return s.CmdSelector(cmd)
	}

	return true
}

func defaultFlagSelect(flag *pflag.Flag) bool {
	if flag.Hidden {
		return false
	}

	if flag.Deprecated != "" {
		return false
	}

	return true
}

// flagSelect returns true if the flag passes default filters and this selector's FlagSelector (if any).
func (s *Selector) flagSelect(flag *pflag.Flag) bool {
	if !defaultFlagSelect(flag) {
		return false
	}

	if s.FlagSelector != nil {
		return s.FlagSelector(flag)
	}

	return true
}

// excludeCmd creates a selector that rejects commands whose path contains any listed phrase.
// Example: excludeCmd("kubectl delete", "admin") excludes "kubectl delete" and "cli admin user".
func excludeCmd(cmds ...string) CmdSelector {
	return func(cmd *cobra.Command) bool {
		for _, phrase := range cmds {
			if strings.Contains(cmd.CommandPath(), phrase) {
				return false
			}
		}

		return true
	}
}
