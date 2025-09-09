package tools

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// TestGeneratorOptions tests various generator configuration options
func TestGeneratorOptions(t *testing.T) {
	t.Run("default generator configuration", func(t *testing.T) {
		gen := NewGenerator()

		// Should have default filters
		assert.Len(t, gen.filters, 3) // Runs(), Hidden(), and Exclude(mcp, help, completion)
	})

	t.Run("generator with custom filters", func(t *testing.T) {
		customFilter := func(cmd *cobra.Command) bool {
			return cmd.Use != "custom-exclude"
		}

		gen := NewGenerator(
			WithFilters(customFilter),
		)

		// Should replace default filters
		assert.Len(t, gen.filters, 1)
	})

	t.Run("generator with additional filter", func(t *testing.T) {
		customFilter := func(_ *cobra.Command) bool {
			return true
		}

		gen := NewGenerator(
			AddFilter(customFilter),
		)

		// Should add to default filters
		assert.Len(t, gen.filters, 4) // 3 defaults + 1 custom
	})

	t.Run("multiple options", func(t *testing.T) {
		handler := func(_ context.Context, _ mcp.CallToolRequest, _ []byte, _ error) (*mcp.CallToolResult, error) {
			return mcp.NewToolResultText("custom"), nil
		}

		gen := NewGenerator(
			WithHandler(handler),
			AddFilter(func(_ *cobra.Command) bool { return true }),
			WithFilters(Allow([]string{"test"})),
		)

		// Handler should be set
		assert.NotNil(t, gen.handler)

		// Filters should be replaced by WithFilters (last option)
		assert.Len(t, gen.filters, 1)
	})
}

// TestFromRootCmdEdgeCases tests edge cases in command tree traversal
func TestFromRootCmdEdgeCases(t *testing.T) {
	t.Run("nil command", func(t *testing.T) {
		gen := NewGenerator()
		tools := gen.fromCmd(nil, "", []Controller{})
		assert.Empty(t, tools)
	})

	t.Run("command with nil subcommands", func(t *testing.T) {
		cmd := &cobra.Command{
			Use: "root",
			Run: func(_ *cobra.Command, _ []string) {},
		}
		// Explicitly set Commands() to return nil
		cmd.Commands() // This initializes the commands slice

		gen := NewGenerator()
		tools := gen.FromRootCmd(cmd)

		assert.Len(t, tools, 1)
		assert.Equal(t, "root", tools[0].Tool.Name)
	})

	t.Run("circular command reference protection", func(t *testing.T) {
		// While Cobra doesn't allow true circular references,
		// this tests that we handle deep nesting properly
		root := &cobra.Command{Use: "root"}
		level1 := &cobra.Command{Use: "level1", Run: func(_ *cobra.Command, _ []string) {}}
		level2 := &cobra.Command{Use: "level2", Run: func(_ *cobra.Command, _ []string) {}}
		level3 := &cobra.Command{Use: "level3", Run: func(_ *cobra.Command, _ []string) {}}

		root.AddCommand(level1)
		level1.AddCommand(level2)
		level2.AddCommand(level3)

		gen := NewGenerator()
		tools := gen.FromRootCmd(root)

		// Should handle all levels without issues
		assert.Len(t, tools, 3)
		expectedNames := []string{"root_level1", "root_level1_level2", "root_level1_level2_level3"}

		var actualNames []string
		for _, tool := range tools {
			actualNames = append(actualNames, tool.Tool.Name)
		}
		assert.ElementsMatch(t, expectedNames, actualNames)
	})
}
