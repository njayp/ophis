package ophis

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestNewCobraToMCPBridge(t *testing.T) {
	rootCmd := &cobra.Command{
		Use:   "test-cli",
		Short: "A test CLI",
	}

	bridge := NewCobraToMCPBridge(rootCmd, "Test App", "1.0.0")

	if bridge.appName != "Test App" {
		t.Errorf("Expected app name 'Test App', got '%s'", bridge.appName)
	}

	if bridge.version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", bridge.version)
	}

	if bridge.rootCmd != rootCmd {
		t.Error("Root command not set correctly")
	}
}

func TestCreateMCPServer(t *testing.T) {
	// Create a simple command structure for testing
	var testFlag string
	var testBool bool
	var testInt int

	rootCmd := &cobra.Command{
		Use:   "test-cli",
		Short: "A test CLI",
	}

	greetCmd := &cobra.Command{
		Use:   "greet [name]",
		Short: "Greet someone",
		Run: func(cmd *cobra.Command, args []string) {
			// Test command
		},
	}

	greetCmd.Flags().StringVar(&testFlag, "message", "hello", "Greeting message")
	greetCmd.Flags().BoolVar(&testBool, "loud", false, "Use loud voice")
	greetCmd.Flags().IntVar(&testInt, "count", 1, "Number of greetings")

	rootCmd.AddCommand(greetCmd)

	// Create bridge and server
	bridge := NewCobraToMCPBridge(rootCmd, "Test App", "1.0.0")
	server := bridge.CreateMCPServer()

	if server == nil {
		t.Fatal("Server creation failed")
	}

	// Verify the server was created and stored
	if bridge.server != server {
		t.Error("Server not stored in bridge")
	}
}

func TestGetCommandDescription(t *testing.T) {
	bridge := &CobraToMCPBridge{}

	tests := []struct {
		name       string
		cmd        *cobra.Command
		parentPath string
		expected   string
	}{
		{
			name: "Command with short description",
			cmd: &cobra.Command{
				Use:   "test",
				Short: "Test command",
			},
			parentPath: "",
			expected:   "Test command",
		},
		{
			name: "Command with long description",
			cmd: &cobra.Command{
				Use:   "test",
				Short: "Test command",
				Long:  "This is a longer description",
			},
			parentPath: "",
			expected:   "Test command\n\nUsage: test\n\nThis is a longer description",
		},
		{
			name: "Command without description",
			cmd: &cobra.Command{
				Use: "test",
			},
			parentPath: "",
			expected:   "Execute the 'test' command",
		},
		{
			name: "Nested command",
			cmd: &cobra.Command{
				Use:   "subcmd",
				Short: "Sub command",
			},
			parentPath: "parent",
			expected:   "Sub command",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := bridge.getCommandDescription(tt.cmd, tt.parentPath)
			if !contains(result, tt.cmd.Short) && tt.cmd.Short != "" {
				t.Errorf("Expected description to contain '%s', got '%s'", tt.cmd.Short, result)
			}
		})
	}
}

func TestSetExecutablePath(t *testing.T) {
	rootCmd := &cobra.Command{
		Use: "test-cli",
	}

	bridge := NewCobraToMCPBridge(rootCmd, "Test App", "1.0.0")
	customPath := "/custom/path/to/binary"

	bridge.SetExecutablePath(customPath)

	if bridge.executablePath != customPath {
		t.Errorf("Expected executable path '%s', got '%s'", customPath, bridge.executablePath)
	}
}

// Helper function to check if a string contains a substring
func contains(str, substr string) bool {
	return len(substr) == 0 || (len(str) >= len(substr) && findSubstring(str, substr))
}

func findSubstring(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
