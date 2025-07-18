package mcp

import (
	"github.com/njayp/ophis/bridge"
	"github.com/spf13/cobra"
)

func validateConfig(config *bridge.Config, cmd *cobra.Command) {
	if config.RootCmd == nil {
		config.RootCmd = cmd.Parent().Parent()
	}

	if config.AppName == "" {
		config.AppName = config.RootCmd.Name()
	}
}
