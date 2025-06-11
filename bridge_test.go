package ophis

import (
	"context"
	"fmt"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/ophis/cmds/basic"
)

func TestExecCommand(t *testing.T) {
	// Create a temporary directory for the test
	cf := basic.NewRootCmd

	// Create a new CobraToMCPBridge instance
	bridge := NewCobraToMCPBridge(cf, "ophis", "0.0.2", nil)

	cmd := cf().Commands()[0]
	fmt.Println(cmd.Use, cmd.Short, cmd.Long, cmd.Name())

	// Execute a command in the temporary directory
	result, err := bridge.executeCommand(context.Background(), cmd, mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("Failed to execute command: %v", err)
	}
	t.Error(result.Content)
}
