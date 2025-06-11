package basic

import (
	"testing"
)

func TestNewRootCmd(t *testing.T) {
	rootCmd := NewRootCmd()

	// Test the hello command
	cmd := rootCmd.Commands()[0]
	if cmd.Flag("greeting").Value.String() != "Hello" {
		t.Errorf("Expected default greeting to be 'Hello', got '%s'", cmd.Flag("greeting").Value.String())
	}
}
