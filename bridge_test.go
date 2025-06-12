package ophis

import (
	"context"
	"fmt"
	"log/slog"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/ophis/cmds/basic"
)

func TestExecCommand(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	// Create a temporary directory for the test
	cf := basic.NewRootCmd

	// Create a new CobraToMCPBridge instance
	bridge := NewCobraToMCPBridge(cf, "ophis", "0.0.0-test", nil)

	cmd := cf().Commands()[0]

	// Execute a command in the temporary directory
	result, err := bridge.executeCommand(context.Background(), cmd, mcp.CallToolRequest{})
	if err != nil {
		t.Error(err.Error())
	}

	content, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Error("content not ok")
	}

	expected := "Hello, World!\n"
	if content.Text != expected {
		t.Error(fmt.Sprintf("wanted %s, got %s", expected, content.Text))
	}
}

func TestFlagDesc(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	// Create a temporary directory for the test
	cf := basic.NewRootCmd

	// Create a new CobraToMCPBridge instance
	bridge := NewCobraToMCPBridge(cf, "ophis", "0.0.0-test", nil)
	t.Error(bridge)
}
