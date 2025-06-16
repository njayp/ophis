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

type HelmCommandFactory struct{}

func (f *HelmCommandFactory) CreateRegistrationCommand() *cobra.Command {
	cmd, err := helmcmd.NewRootCmd(nil, nil)
	if err != nil {
		panic(err)
	}

	return cmd
}

func (f *HelmCommandFactory) CreateCommand() (*cobra.Command, bridge.CommandExecFunc) {
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
