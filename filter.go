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
		ExcludeFilter([]string{cfgmgr.MCPCommandName, "help", "completion"}),
	}
}

// ExcludeFilter creates a filter that rejects commands whose path contains any listed phrase.
// Example: ExcludeFilter([]string{"kubectl delete", "admin"}) excludes "kubectl delete" and "cli admin user".
func ExcludeFilter(list []string) Filter {
	return func(cmd *cobra.Command) bool {
		for _, phrase := range list {
			if strings.Contains(cmd.CommandPath(), phrase) {
				slog.Debug("excluding command by exclude list", "command_path", cmd.CommandPath(), "phrase", phrase)
				return false
			}
		}

		return true
	}
}

// AllowFilter creates a filter that only accepts commands whose path contains a listed phrase.
// Example: AllowFilter([]string{"get", "helm list"}) includes "kubectl get pods" and "helm list".
func AllowFilter(list []string) Filter {
	return func(cmd *cobra.Command) bool {
		for _, phrase := range list {
			if strings.Contains(cmd.CommandPath(), phrase) {
				return true
			}
		}

		slog.Debug("excluding command by allow list", "command_path", cmd.CommandPath(), "allow_list", list)
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
