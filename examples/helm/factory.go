// Package main provides an example MCP server that exposes Helm commands.
// This demonstrates how to use ophis to convert the Helm CLI into an MCP server.
package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/ophis/bridge"
	"github.com/spf13/cobra"
	helmcmd "helm.sh/helm/v4/pkg/cmd"
)

// HelmCommandFactory implements the bridge.CommandFactory interface for Helm commands.
type HelmCommandFactory struct{}

// RegistrationCommand creates a Helm command tree for MCP tool registration.
func (f *HelmCommandFactory) RegistrationCommand() *cobra.Command {
	cmd, err := helmcmd.NewRootCmd(nil, nil)
	if err != nil {
		panic(err)
	}

	return cmd
}

// New creates a fresh Helm command instance and its execution function.
func (f *HelmCommandFactory) New() (*cobra.Command, bridge.CommandExecFunc) {
	var output strings.Builder

	cmd, err := helmcmd.NewRootCmd(&output, nil)
	if err != nil {
		panic(err)
	}

	exec := func(ctx context.Context) *mcp.CallToolResult {
		err := cmd.ExecuteContext(ctx)
		if err != nil {
			return mcp.NewToolResultErrorFromErr(fmt.Sprintf("Command %s failed:", cmd.CommandPath()), err)
		}
		return mcp.NewToolResultText(output.String())
	}

	return cmd, exec
}
