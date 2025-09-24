package ophis

import (
	"log/slog"
	"strings"

	"github.com/njayp/ophis/internal/cfgmgr"
	"github.com/spf13/cobra"
)

// Filter determines if a command should become an MCP tool.
// Returns true to include the command.
type Filter func(*cobra.Command) bool

func defaultFilters() []Filter {
	return []Filter{
		runsFilter(),
		hiddenFilter(),
		depreciatedFilter(),
		ExcludeFilter(cfgmgr.MCPCommandName, "help", "completion"),
	}
}

// ExcludeFilter creates a filter that rejects commands whose path contains any listed phrase.
// Example: ExcludeFilter("kubectl delete", "admin") excludes "kubectl delete" and "cli admin user".
func ExcludeFilter(cmds ...string) Filter {
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

// AllowFilter creates a filter that only accepts commands whose path contains a listed phrase.
// Example: AllowFilter("get", "helm list") includes "kubectl get pods" and "helm list".
func AllowFilter(cmds ...string) Filter {
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

// hiddenFilter creates a filter that excludes hidden commands.
func hiddenFilter() Filter {
	return func(cmd *cobra.Command) bool {
		if cmd.Hidden {
			slog.Debug("excluding hidden command", "command", cmd.Name())
		}
		return !cmd.Hidden
	}
}

// depreciatedFilter creates a filter that excludes hidden commands.
func depreciatedFilter() Filter {
	return func(cmd *cobra.Command) bool {
		deprecated := cmd.Deprecated != ""
		if deprecated {
			slog.Debug("excluding depreciated command", "command", cmd.Name())
		}

		return !deprecated
	}
}

// runsFilter creates a filter that excludes non-runnable commands.
func runsFilter() Filter {
	return func(cmd *cobra.Command) bool {
		noop := cmd.Run == nil && cmd.RunE == nil && cmd.PreRun == nil && cmd.PreRunE == nil
		if noop {
			slog.Debug("excluding command without run or pre-run function", "command", cmd.CommandPath())
		}
		return !noop
	}
}
