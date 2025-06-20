// Package main provides an example MCP server that exposes make commands.
// This demonstrates how to use njayp/ophis to convert a make-based build system into an MCP server.
package main

import (
	"context"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/njayp/ophis/bridge"
	"github.com/njayp/ophis/tools"
	"github.com/spf13/cobra"
)

// CommandFactory implements the bridge.CommandFactory interface for make commands.
type CommandFactory struct {
	rootCmd *cobra.Command
}

// Tools returns the list of MCP tools from the command tree.
func (f *CommandFactory) Tools() []tools.Tool {
	return tools.FromRootCmd(f.rootCmd)
}

// New creates a fresh command instance and its execution function.
func (f *CommandFactory) New() (*cobra.Command, bridge.CommandExecFunc) {
	var output strings.Builder
	rootCmd := createMakeCommands()
	rootCmd.SetOut(&output)
	rootCmd.SetErr(&output)

	execFunc := func(ctx context.Context, cmd *cobra.Command) *mcp.CallToolResult {
		err := cmd.ExecuteContext(ctx)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("Failed to execute command", err)
		}
		return mcp.NewToolResultText(output.String())
	}

	return rootCmd, execFunc
}
