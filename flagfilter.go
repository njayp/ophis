package ophis

import (
	"log/slog"
	"strings"

	"github.com/spf13/pflag"
)

// FlagFilter determines if a flag should be included in MCP tools.
// Returns true to include the flag.
type FlagFilter func(*pflag.Flag) bool

func defaultFlagFilters() []FlagFilter {
	return []FlagFilter{
		hiddenFlagFilter(),
		depreciatedFlagFilter(),
	}
}

// ExcludeFlagFilter creates a filter that rejects commands whose path contains any listed phrase.
// Example: ExcludeFilter("kubectl delete", "admin") excludes "kubectl delete" and "cli admin user".
func ExcludeFlagFilter(cmds ...string) FlagFilter {
	return func(flag *pflag.Flag) bool {
		for _, phrase := range cmds {
			if strings.Contains(flag.Name, phrase) {
				slog.Debug("excluding command by exclude list", "command_path", flag.Name, "phrase", phrase)
				return false
			}
		}

		return true
	}
}

// AllowFlagFilter creates a filter that only accepts commands whose path contains a listed phrase.
// Example: AllowFilter("get", "helm list") includes "kubectl get pods" and "helm list".
func AllowFlagFilter(cmds ...string) FlagFilter {
	return func(flag *pflag.Flag) bool {
		for _, phrase := range cmds {
			if strings.Contains(flag.Name, phrase) {
				return true
			}
		}

		slog.Debug("excluding command by allow list", "command_path", flag.Name, "allow_list", cmds)
		return false
	}
}

// hiddenFilter creates a filter that excludes hidden commands.
func hiddenFlagFilter() FlagFilter {
	return func(flag *pflag.Flag) bool {
		if flag.Hidden {
			slog.Debug("excluding hidden flag", "command", flag.Name)
		}
		return !flag.Hidden
	}
}

// depreciatedFilter creates a filter that excludes hidden commands.
func depreciatedFlagFilter() FlagFilter {
	return func(flag *pflag.Flag) bool {
		deprecated := flag.Deprecated != ""
		if deprecated {
			slog.Debug("excluding depreciated command", "command", flag.Name)
		}

		return !deprecated
	}
}
