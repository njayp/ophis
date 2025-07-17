package tools

import (
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
		return !slices.Contains(list, cmd.Name())
	}
}

// Allow adds a filter to include only subcommands that match the provided list.
// It checks if the command path contains any of the specified white-listed command names.
// Therefore, it only works for first-level subcommands.
func Allow(list []string) Filter {
	return func(cmd *cobra.Command) bool {
		for _, name := range list {
			if strings.Contains(cmd.CommandPath(), name) {
				return true
			}
		}

		return false
	}
}

// Hidden returns a filter that excludes hidden commands from the generated tools.
func Hidden() Filter {
	return func(cmd *cobra.Command) bool {
		return !cmd.Hidden
	}
}
