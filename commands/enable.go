package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

type EnableCommandFlags struct {
	ConfigPath string
	LogLevel   string
	LogFile    string
}

func EnableCommand() *cobra.Command {
	enableFlags := &EnableCommandFlags{}
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start MCP (Model Context Protocol) server",
		Long: fmt.Sprintf(`Start an MCP server that exposes this application's commands to MCP clients.

The MCP server will expose all available commands as tools that can be called
by AI assistants and other MCP-compatible clients.`),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO add self to config files
			return nil
		},
	}

	// Add flags
	flags := cmd.Flags()
	flags.StringVar(&enableFlags.LogLevel, "log-level", "", "Log level (debug, info, warn, error)")
	flags.StringVar(&enableFlags.LogFile, "log-file", "", "Path to log file (default: user cache)")
	flags.StringVar(&enableFlags.ConfigPath, "config-path", "", "Path to config file")
	return cmd
}

func DisableCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start MCP (Model Context Protocol) server",
		Long: fmt.Sprintf(`Start an MCP server that exposes this application's commands to MCP clients.

The MCP server will expose all available commands as tools that can be called
by AI assistants and other MCP-compatible clients.`),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO
			return nil
		},
	}

	return cmd
}
