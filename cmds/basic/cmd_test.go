package basic

import (
	"testing"

	"github.com/spf13/pflag"
)

func TestNewRootCmd(t *testing.T) {
	rootCmd := NewRootCmd()

	// Test the hello command
	helloCmd := rootCmd.Commands()[0]
	helloCmd.Flags().VisitAll(func(f *pflag.Flag) {
		if f.Name == "greeting" {
			if f.Value.String() != "Hello" {
				t.Errorf("Expected default greeting to be 'Hello', got '%s'", f.Value.String())
			}
		}
	})
}
