package main

import (
	"os"

	"github.com/ophis/bridge"
	"github.com/ophis/commands"
	"github.com/spf13/cobra"
)

// Configuration constants
const (
	AppName    = "helm"
	AppVersion = "0.0.1"
)

func main() {
	// Create the root command
	rootCmd := &cobra.Command{
		Use:   "helm",
		Short: "The Helm package manager for Kubernetes",
		Long:  `The Helm package manager for Kubernetes with MCP support`,
	}

	// Add the MCP command as a subcommand
	mcpConfig := &bridge.MCPCommandConfig{
		AppName:    AppName,
		AppVersion: AppVersion,
	}
	rootCmd.AddCommand(commands.MCPCommand(&HelmCommandFactory{}, mcpConfig))

	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
