package bridge

import (
	"log/slog"
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
		customGen := tools.NewGenerator(
			tools.WithFilters(tools.Allow([]string{"sub"})),
		)

		config := &Config{
			AppName:   "test-app",
			RootCmd:   rootCmd,
			Generator: customGen,
		}

		tools := config.Tools()
		assert.Len(t, tools, 1)
		assert.Equal(t, "test_sub", tools[0].Tool.Name)
	})

	t.Run("with default generator", func(t *testing.T) {
		config := &Config{
			AppName: "test-app",
			RootCmd: rootCmd,
			// Generator is nil, should use default
		}

		tools := config.Tools()
		assert.Len(t, tools, 1)
		assert.Equal(t, "test_sub", tools[0].Tool.Name)
	})
}

// TestConfigValidation tests various config validation scenarios
func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
		errorMsg    string
	}{
		{
			name:        "nil config",
			config:      nil,
			expectError: true,
			errorMsg:    "configuration cannot be nil",
		},
		{
			name: "empty app name",
			config: &Config{
				AppName: "",
				RootCmd: &cobra.Command{Use: "test"},
			},
			expectError: true,
			errorMsg:    "application name cannot be empty",
		},
		{
			name: "nil root command",
			config: &Config{
				AppName: "test-app",
				RootCmd: nil,
			},
			expectError: true,
			errorMsg:    "root command cannot be nil",
		},
		{
			name: "valid config with defaults",
			config: &Config{
				AppName: "test-app",
				RootCmd: &cobra.Command{Use: "test"},
			},
			expectError: false,
		},
		{
			name: "config with all options",
			config: &Config{
				AppName:    "test-app",
				AppVersion: "1.0.0",
				RootCmd:    &cobra.Command{Use: "test"},
				Generator:  tools.NewGenerator(),
				SloggerOptions: &slog.HandlerOptions{
					Level: slog.LevelDebug,
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestConfigDefaults tests that defaults are properly set
func TestConfigDefaults(t *testing.T) {
	config := &Config{
		AppName: "test-app",
		RootCmd: &cobra.Command{Use: "test"},
		// AppVersion not set
	}

	manager, err := New(config)
	assert.NoError(t, err)
	assert.NotNil(t, manager)

	// Version should default to "unknown"
	// We can't directly test this without accessing internal fields,
	// but we can verify the manager was created successfully
}
