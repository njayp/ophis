package tools

import (
	"log/slog"
	"slices"
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
func Exclude(list []string) Filter {
	return func(cmd *cobra.Command) bool {
		excluded := slices.Contains(list, cmd.Name())
		if excluded {
			slog.Debug("excluding command by name", "command", cmd.Name(), "exclude_list", list)
		}
		return !excluded
	}
}

// Allow adds a filter to include only subcommands that match the provided list.
// It checks if the command path contains any of the specified white-listed command names.
// Therefore, it only works for first-level subcommands.
func Allow(list []string) Filter {
	return func(cmd *cobra.Command) bool {
		for _, name := range list {
			if strings.Contains(cmd.CommandPath(), name) {
				slog.Debug("allowing command by path", "command", cmd.CommandPath(), "matched", name)
				return true
			}
		}
		slog.Debug("filtering out command not in allow list", "command", cmd.CommandPath(), "allow_list", list)
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
