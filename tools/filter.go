package tools

import (
	"log/slog"
	"strings"

	"github.com/spf13/cobra"
)

// Filter is a function type used by the Generator to filter commands.
// It returns true if the command should be included in the generated tools.
type Filter func(*cobra.Command) bool

// WithFilters sets custom filters for the generator.
func WithFilters(filters ...Filter) GeneratorOption {
	return func(g *Generator) {
		g.filters = filters
	}
}

// AddFilter adds a custom filter function to the generator.
func AddFilter(filter Filter) GeneratorOption {
	return func(g *Generator) {
		g.filters = append(g.filters, filter)
	}
}

// Exclude adds a filter to exclude listed command names from the generated tools.
// E.g., Exclude([]string{"delete", "user test"}) will exclude any command whose
// path contains "delete" or "user test".
func Exclude(list []string) Filter {
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

// Allow adds a filter to include only subcommands that match the provided list.
// E.g., Allow([]string{"get", "user info"}) will only include any command whose
// path contains "get" or "user info".
func Allow(list []string) Filter {
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

// Hidden returns a filter that excludes hidden commands from the generated tools.
func Hidden() Filter {
	return func(cmd *cobra.Command) bool {
		if cmd.Hidden {
			slog.Debug("excluding hidden command", "command", cmd.Name())
		}
		return !cmd.Hidden
	}
}

// Runs returns a filter that excludes commands without a Run or PreRun function.
func Runs() Filter {
	return func(cmd *cobra.Command) bool {
		noop := cmd.Run == nil && cmd.RunE == nil && cmd.PreRun == nil && cmd.PreRunE == nil
		if noop {
			slog.Debug("excluding command without run or pre-run function", "command", cmd.CommandPath())
		}
		return !noop
	}
}
