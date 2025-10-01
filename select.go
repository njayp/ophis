package ophis

import (
	"slices"
	"strings"

	"github.com/njayp/ophis/internal/bridge"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// CmdSelector determines if a command should become an MCP tool.
// Return true to include the command as a tool.
// Note: Basic safety filters (hidden, deprecated, non-runnable) are always applied first.
// Commands are tested against selectors in order; the first matching selector wins.
type CmdSelector bridge.CmdSelector

// FlagSelector determines if a flag should be included in an MCP tool.
// Return true to include the flag.
// Note: Hidden and deprecated flags are always excluded regardless of this selector.
// This selector is only applied to commands that match the associated CmdSelector.
type FlagSelector bridge.FlagSelector

// PreRunFunc is middleware hook that runs before each tool call
// Return a cancelled context to prevent execution.
// Common uses: add timeouts, rate limiting, auth checks, metrics.
type PreRunFunc bridge.PreRunFunc

// PostRunFunc is middleware hook that runs after each tool call
// Common uses: error handling, response filtering, metrics collection.
type PostRunFunc bridge.PostRunFunc

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

// ExcludeCmd creates a selector that rejects commands whose path contains any listed phrase.
// Example: ExcludeCmd("kubectl delete", "admin") excludes "kubectl delete" and "cli admin user".
func ExcludeCmd(cmds ...string) CmdSelector {
	return func(cmd *cobra.Command) bool {
		for _, phrase := range cmds {
			if strings.Contains(cmd.CommandPath(), phrase) {
				return false
			}
		}

		return true
	}
}

// AllowCmd creates a selector that only accepts commands whose path contains a listed phrase.
// Example: AllowCmd("get", "helm list") includes "kubectl get pods" and "helm list".
func AllowCmd(cmds ...string) CmdSelector {
	return func(cmd *cobra.Command) bool {
		for _, phrase := range cmds {
			if strings.Contains(cmd.CommandPath(), phrase) {
				return true
			}
		}

		return false
	}
}

// ExcludeFlag creates a selector that rejects flags whose name is listed.
// Example: ExcludeFlag("color", "kubeconfig") excludes flags named "color" and "kubeconfig".
func ExcludeFlag(names ...string) FlagSelector {
	return func(flag *pflag.Flag) bool {
		return !slices.Contains(names, flag.Name)
	}
}

// AllowFlag creates a selector that only accepts flags whose name is listed.
// Example: AllowFlag("namespace", "output") includes only flags named "namespace" and "output".
func AllowFlag(names ...string) FlagSelector {
	return func(flag *pflag.Flag) bool {
		return slices.Contains(names, flag.Name)
	}
}

// NoFlags is a FlagSelector that excludes all flags.
func NoFlags(_ *pflag.Flag) bool {
	return false
}
