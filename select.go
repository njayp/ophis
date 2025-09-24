package ophis

import (
	"log/slog"
	"strings"

	"github.com/njayp/ophis/internal/cfgmgr"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// CmdSelector determines if a command should become an MCP tool.
// Return true to include the command as a tool.
// Commands are tested against selectors in order; the first matching selector wins.
type CmdSelector func(*cobra.Command) bool

// FlagSelector determines if a flag should be included in an MCP tool.
// Return true to include the flag.
// This selector is only applied to commands that match the associated CmdSelector.
type FlagSelector func(*pflag.Flag) bool

// Selector contains selectors for filtering commands and flags.
// When multiple selectors are configured, they are evaluated in order.
// The first selector whose CmdSelect matches a command is used,
// and its FlagSelect determines which flags are included for that command.
//
// This allows fine-grained control, such as:
//   - Exposing different flags for different command groups
//   - Applying stricter flag filtering to dangerous commands
//   - Having a default catch-all selector with common flag exclusions
type Selector struct {
	// CmdSelector determines if this selector applies to a command.
	// If nil, defaults to accepting commands that are runnable, visible, and not deprecated.
	CmdSelector CmdSelector
	// FlagSelector determines which flags to include for commands matched by CmdSelect.
	// If nil, defaults to including all visible, non-deprecated flags.
	FlagSelector FlagSelector
}

func (s *Selector) cmdSelect(cmd *cobra.Command) bool {
	if !defaultCmdSelect(cmd) {
		return false
	}

	if s.CmdSelector != nil && !s.CmdSelector(cmd) {
		return false
	}

	return true
}

// ExcludeCmd creates a selector that rejects commands whose path contains any listed phrase.
// Example: ExcludeCmd("kubectl delete", "admin") excludes "kubectl delete" and "cli admin user".
func ExcludeCmd(cmds ...string) CmdSelector {
	return func(cmd *cobra.Command) bool {
		for _, phrase := range cmds {
			if strings.Contains(cmd.CommandPath(), phrase) {
				slog.Debug("excluding command by exclude list", "command_path", cmd.CommandPath(), "phrase", phrase)
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

		slog.Debug("excluding command by allow list", "command_path", cmd.CommandPath(), "allow_list", cmds)
		return false
	}
}

// ExcludeFlag creates a selector that rejects flags whose name contains any listed phrase.
// Example: ExcludeFlag("color", "kubeconfig") excludes flags named "color" and "kubeconfig".
func ExcludeFlag(names ...string) FlagSelector {
	return func(flag *pflag.Flag) bool {
		for _, phrase := range names {
			if strings.Contains(flag.Name, phrase) {
				slog.Debug("excluding flag by exclude list", "flag_name", flag.Name, "phrase", phrase)
				return false
			}
		}

		return true
	}
}

// AllowFlag creates a selector that only accepts flags whose name contains a listed phrase.
// Example: AllowFlag("namespace", "output") includes only flags named "namespace" and "output".
func AllowFlag(names ...string) FlagSelector {
	return func(flag *pflag.Flag) bool {
		for _, phrase := range names {
			if strings.Contains(flag.Name, phrase) {
				return true
			}
		}

		slog.Debug("excluding flag by allow list", "flag_name", flag.Name, "allow_list", names)
		return false
	}
}

func defaultCmdSelect(c *cobra.Command) bool {
	if c.Hidden {
		slog.Debug("excluding hidden command", "command", c.CommandPath())
		return false
	}

	if c.Deprecated != "" {
		slog.Debug("excluding depreciated command", "command", c.CommandPath())
		return false
	}

	if c.Run == nil && c.RunE == nil && c.PreRun == nil && c.PreRunE == nil {
		slog.Debug("excluding command without run or pre-run function", "command", c.CommandPath())
		return false
	}

	return ExcludeCmd(cfgmgr.MCPCommandName, "help", "completion")(c)
}
