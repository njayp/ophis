package ophis

import (
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// AllowCmdsContaining creates a selector that only accepts commands whose path contains a listed phrase.
// Example: AllowCmdsContaining("get", "helm list") includes "kubectl get pods" and "helm list".
func AllowCmdsContaining(substrings ...string) CmdSelector {
	return func(cmd *cobra.Command) bool {
		for _, s := range substrings {
			if strings.Contains(cmd.CommandPath(), s) {
				return true
			}
		}

		return false
	}
}

// ExcludeCmdsContaining creates a selector that rejects commands whose path contains any listed phrase.
// Example: ExcludeCmdsContaining("kubectl delete", "admin") excludes "kubectl delete" and "cli admin user".
func ExcludeCmdsContaining(substrings ...string) CmdSelector {
	selector := AllowCmdsContaining(substrings...)
	return func(cmd *cobra.Command) bool {
		return !selector(cmd)
	}
}

// AllowCmds creates a selector that only accepts commands whose path is listed.
// Example: AllowCmds("kubectl get", "helm list") includes only those exact commands.
func AllowCmds(cmds ...string) CmdSelector {
	return func(cmd *cobra.Command) bool {
		return slices.Contains(cmds, cmd.CommandPath())
	}
}

// ExcludeCmds creates a selector that rejects commands whose path is listed.
// Example: ExcludeCmds("kubectl delete", "helm uninstall") excludes those exact commands.
func ExcludeCmds(cmds ...string) CmdSelector {
	return func(cmd *cobra.Command) bool {
		return !slices.Contains(cmds, cmd.CommandPath())
	}
}

// AllowFlags creates a selector that only accepts flags whose name is listed.
// Example: AllowFlags("namespace", "output") includes only flags named "namespace" and "output".
func AllowFlags(names ...string) FlagSelector {
	return func(flag *pflag.Flag) bool {
		return slices.Contains(names, flag.Name)
	}
}

// ExcludeFlags creates a selector that rejects flags whose name is listed.
// Example: ExcludeFlags("color", "kubeconfig") excludes flags named "color" and "kubeconfig".
func ExcludeFlags(names ...string) FlagSelector {
	return func(flag *pflag.Flag) bool {
		return !slices.Contains(names, flag.Name)
	}
}

// NoFlags is a FlagSelector that excludes all flags.
func NoFlags(_ *pflag.Flag) bool {
	return false
}
