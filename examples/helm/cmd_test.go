package main

import (
	"context"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/ophis"
)

func TestHelmCommandFactoryImplementation(t *testing.T) {
	factory := &HelmCommandFactory{}

	t.Run("implements CommandFactory interface", func(t *testing.T) {
		// This will fail to compile if HelmCommandFactory doesn't implement CommandFactory
		var _ ophis.CommandFactory = factory
	})

	t.Run("registration command has expected structure", func(t *testing.T) {
		cmd := factory.CreateRegistrationCommand()

		// Check that it's a helm command
		if !strings.Contains(cmd.Use, "helm") && cmd.Name() != "helm" {
			t.Error("Expected helm command")
		}

		// Check that it has subcommands (helm has many)
		if len(cmd.Commands()) == 0 {
			t.Error("Expected helm to have subcommands")
		}

		// Verify some common helm commands exist
		commandNames := make(map[string]bool)
		for _, subcmd := range cmd.Commands() {
			commandNames[subcmd.Name()] = true
		}

		expectedCommands := []string{"install", "upgrade", "list", "uninstall"}
		for _, expected := range expectedCommands {
			if !commandNames[expected] {
				t.Errorf("Expected to find helm command: %s", expected)
			}
		}
	})

	t.Run("execution command creates fresh instance", func(t *testing.T) {
		cmd1, _ := factory.CreateCommand()
		cmd2, _ := factory.CreateCommand()

		// These should be different instances
		if cmd1 == cmd2 {
			t.Error("Expected different command instances")
		}
	})

	t.Run("execution function returns proper result", func(t *testing.T) {
		_, exec := factory.CreateCommand()

		if exec == nil {
			t.Fatal("Expected execution function")
		}

		// Execute with context
		result := exec(context.Background())

		if result == nil {
			t.Fatal("Expected execution result")
		}

		// The result should either be successful (with help output) or an error
		// depending on how helm handles no arguments
		if result.IsError {
			t.Logf("Helm execution returned error as expected: %v", result.Content)
		} else {
			t.Logf("Helm execution succeeded: %v", result.Content)
		}
	})
}

func TestHelmCommandExecution(t *testing.T) {
	factory := &HelmCommandFactory{}

	t.Run("help command", func(t *testing.T) {
		cmd, exec := factory.CreateCommand()

		// Set args for help
		cmd.SetArgs([]string{"help"})

		result := exec(context.Background())

		if result.IsError {
			t.Errorf("Help command should not error: %v", result.Content)
		}

		// Check that output contains expected help text
		if len(result.Content) > 0 {
			if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
				if !strings.Contains(textContent.Text, "helm") {
					t.Error("Expected help output to contain 'helm'")
				}
			}
		}
	})

	t.Run("version command", func(t *testing.T) {
		cmd, exec := factory.CreateCommand()

		// Set args for version
		cmd.SetArgs([]string{"version"})

		result := exec(context.Background())

		if result.IsError {
			t.Errorf("Version command should not error: %v", result.Content)
		}

		// Check that output contains version information
		if len(result.Content) > 0 {
			if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
				output := textContent.Text
				if !strings.Contains(output, "version") && !strings.Contains(output, "Version") {
					t.Error("Expected version output to contain version information")
				}
			}
		}
	})
}

// TestWithOutput verifies that command output is properly captured
func TestWithOutput(t *testing.T) {
	factory := &HelmCommandFactory{}

	// Test that multiple command creations produce independent output
	cmd1, exec1 := factory.CreateCommand()
	cmd2, exec2 := factory.CreateCommand()

	cmd1.SetArgs([]string{"help"})
	cmd2.SetArgs([]string{"version"})

	result1 := exec1(context.Background())
	result2 := exec2(context.Background())

	// Both should succeed
	if result1.IsError {
		t.Errorf("First command failed: %v", result1.Content)
	}
	if result2.IsError {
		t.Errorf("Second command failed: %v", result2.Content)
	}

	// Results should be different (help vs version output)
	if len(result1.Content) > 0 && len(result2.Content) > 0 {
		text1 := ""
		text2 := ""

		if tc1, ok := result1.Content[0].(*mcp.TextContent); ok {
			text1 = tc1.Text
		}
		if tc2, ok := result2.Content[0].(*mcp.TextContent); ok {
			text2 = tc2.Text
		}

		if text1 == text2 && text1 != "" {
			t.Error("Expected different outputs from help and version commands")
		}
	}
}
