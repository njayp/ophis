package bridge

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/spf13/cobra"
)

func TestGetCommandDescription(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))
	bridge := &CobraToMCPBridge{logger: logger}

	tests := []struct {
		name       string
		cmd        *cobra.Command
		parentPath string
		contains   []string // Strings that should be in the description
	}{
		{
			name: "command with short description",
			cmd: &cobra.Command{
				Use:   "test",
				Short: "Short description",
			},
			parentPath: "",
			contains:   []string{"Short description"},
		},
		{
			name: "command with long description",
			cmd: &cobra.Command{
				Use:   "test",
				Short: "Short description",
				Long:  "This is a longer description with more details",
			},
			parentPath: "",
			contains:   []string{"Short description", "longer description"},
		},
		{
			name: "command with usage",
			cmd: &cobra.Command{
				Use:   "test [flags]",
				Short: "Test command",
			},
			parentPath: "",
			contains:   []string{"Test command", "Usage: test [flags]"},
		},
		{
			name: "command without description",
			cmd: &cobra.Command{
				Use: "test",
			},
			parentPath: "parent",
			contains:   []string{"Execute the 'parent test' command"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			desc := bridge.getCommandDescription(tt.cmd, tt.parentPath)

			for _, expected := range tt.contains {
				if !containsIgnoreCase(desc, expected) {
					t.Errorf("Expected description to contain '%s', got: %s", expected, desc)
				}
			}
		})
	}
}

func TestFlagMapFromCmd(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	bridge := &CobraToMCPBridge{logger: slog.Default()}

	cmd := &cobra.Command{
		Use: "test",
	}

	// Add various flag types
	cmd.Flags().String("string-flag", "", "A string flag")
	cmd.Flags().Bool("bool-flag", false, "A boolean flag")
	cmd.Flags().Int("int-flag", 0, "An integer flag")
	cmd.Flags().StringSlice("slice-flag", nil, "A string slice flag")

	// Add a hidden flag
	hiddenFlag := cmd.Flags().String("hidden-flag", "", "A hidden flag")
	cmd.Flags().MarkHidden("hidden-flag")
	_ = hiddenFlag

	// Add inherited flags
	parentCmd := &cobra.Command{Use: "parent"}
	parentCmd.PersistentFlags().String("inherited-flag", "", "An inherited flag")
	parentCmd.AddCommand(cmd)

	flagMap := bridge.flagMapFromCmd(cmd)

	// Check that expected flags are present
	expectedFlags := []string{"string-flag", "bool-flag", "int-flag", "slice-flag", "inherited-flag"}
	for _, flagName := range expectedFlags {
		if _, exists := flagMap[flagName]; !exists {
			t.Errorf("Expected flag '%s' to be in flag map", flagName)
		}
	}

	// Check that hidden flag is not present
	if _, exists := flagMap["hidden-flag"]; exists {
		t.Error("Hidden flag should not be in flag map")
	}

	// Verify flag properties
	if stringFlag, ok := flagMap["string-flag"].(map[string]string); ok {
		if stringFlag["type"] != "string" {
			t.Errorf("Expected string flag type to be 'string', got '%s'", stringFlag["type"])
		}
		if stringFlag["description"] != "A string flag" {
			t.Errorf("Expected string flag description to be 'A string flag', got '%s'", stringFlag["description"])
		}
	} else {
		t.Error("String flag should be a map[string]string")
	}
}

func TestArgsDescFromCmd(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))
	bridge := &CobraToMCPBridge{logger: logger}

	tests := []struct {
		name     string
		cmd      *cobra.Command
		contains []string
	}{
		{
			name: "command with usage",
			cmd: &cobra.Command{
				Use: "test [file]",
			},
			contains: []string{"Space-separated positional arguments", "Usage: test [file]"},
		},
		{
			name: "command with args validation",
			cmd: &cobra.Command{
				Use:  "test",
				Args: cobra.ExactArgs(1),
			},
			contains: []string{"Space-separated positional arguments", "argument requirements"},
		},
		{
			name: "simple command",
			cmd: &cobra.Command{
				Use: "test",
			},
			contains: []string{"Space-separated positional arguments"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			desc := bridge.argsDescFromCmd(tt.cmd)

			for _, expected := range tt.contains {
				if !containsIgnoreCase(desc, expected) {
					t.Errorf("Expected args description to contain '%s', got: %s", expected, desc)
				}
			}
		})
	}
}

func TestFlagToolOption(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("test-flag", "", "Test flag description")

	option := flagToolOption(cmd.Flag("test-flag"))

	if option["type"] != "string" {
		t.Errorf("Expected flag type to be 'string', got '%s'", option["type"])
	}

	if option["description"] != "Test flag description" {
		t.Errorf("Expected flag description to be 'Test flag description', got '%s'", option["description"])
	}

	// Test flag without description
	cmd.Flags().String("no-desc", "", "")
	noDescOption := flagToolOption(cmd.Flag("no-desc"))

	expectedDesc := "Flag: no-desc"
	if noDescOption["description"] != expectedDesc {
		t.Errorf("Expected default description '%s', got '%s'", expectedDesc, noDescOption["description"])
	}
}

// Helper function for case-insensitive string matching
func containsIgnoreCase(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}

// Mock MCP server for testing
type mockMCPServer struct {
	addToolFunc func(toolName string)
}

func (m *mockMCPServer) AddTool(tool interface{}, handler interface{}) {
	if m.addToolFunc != nil {
		// Extract tool name if possible, otherwise use a default
		m.addToolFunc("mock-tool")
	}
}
