package bridge

import (
	"testing"

	"github.com/njayp/ophis/tools"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// TestConfigTools tests the Tools() method of Config
func TestConfigTools(t *testing.T) {
	rootCmd := &cobra.Command{
		Use:   "test",
		Short: "Test CLI",
	}

	subCmd := &cobra.Command{
		Use:   "sub",
		Short: "Subcommand",
		Run:   func(_ *cobra.Command, _ []string) {},
	}

	rootCmd.AddCommand(subCmd)

	t.Run("with custom generator", func(t *testing.T) {
		customGen := []tools.GeneratorOption{tools.WithFilters(tools.Allow([]string{"sub"}))}

		config := &Config{
			RootCmd:          rootCmd,
			GeneratorOptions: customGen,
		}

		tools := config.Tools()
		assert.Len(t, tools, 1)
		assert.Equal(t, "test_sub", tools[0].Tool.Name)
	})

	t.Run("with default generator", func(t *testing.T) {
		config := &Config{
			RootCmd: rootCmd,
			// Generator is nil, should use default
		}

		tools := config.Tools()
		assert.Len(t, tools, 1)
		assert.Equal(t, "test_sub", tools[0].Tool.Name)
	})
}
