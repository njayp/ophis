package ophis

import (
	"log/slog"
	"strings"

	"github.com/njayp/ophis/internal/cfgmgr"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// CmdSelector determines if a command should become an MCP tool.
// Return true to include the command.
type CmdSelector func(*cobra.Command) bool

// FlagSelector determines if a flag should be included in an MCP tool.
// Return true to include the flag.
type FlagSelector func(*pflag.Flag) bool

// Selector
type Selector struct {
	// allowed commands
	CmdSelect CmdSelector
	// allowed flags on these commands
	FlagSelect FlagSelector
}

// ExcludeCmd creates a filter that rejects commands whose path contains any listed phrase.
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

// AllowCmd creates a filter that only accepts commands whose path contains a listed phrase.
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

// ExcludeFlag creates a filter that rejects commands whose path contains any listed phrase.
// Example: ExcludeFilter("kubectl delete", "admin") excludes "kubectl delete" and "cli admin user".
func ExcludeFlag(names ...string) FlagSelector {
	return func(flag *pflag.Flag) bool {
		for _, phrase := range names {
			if strings.Contains(flag.Name, phrase) {
				slog.Debug("excluding command by exclude list", "command_path", flag.Name, "phrase", phrase)
				return false
			}
		}

		return true
	}
}

// AllowFlag creates a filter that only accepts commands whose path contains a listed phrase.
// Example: AllowFilter("get", "helm list") includes "kubectl get pods" and "helm list".
func AllowFlag(names ...string) FlagSelector {
	return func(flag *pflag.Flag) bool {
		for _, phrase := range names {
			if strings.Contains(flag.Name, phrase) {
				return true
			}
		}

		slog.Debug("excluding command by allow list", "command_path", flag.Name, "allow_list", names)
		return false
	}
}

func defaultSelect() []Selector {
	return []Selector{
		{
			CmdSelect:  defaultCmdSelect(),
			FlagSelect: defaultFlagSelect(),
		},
	}
}

func defaultCmdSelect() CmdSelector {
	return func(c *cobra.Command) bool {
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

		if ExcludeCmd(cfgmgr.MCPCommandName, "help", "completion")(c) {
			return false
		}

		return true
	}
}

func defaultFlagSelect() FlagSelector {
	return func(f *pflag.Flag) bool {
		return !f.Hidden && f.Deprecated == ""
	}
}
