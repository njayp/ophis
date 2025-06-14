package commands

import (
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
		Short: "Enable the MCP server",
		Long:  `Enable the MCP server by adding it to claude's MCP config file`,
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
		Short: "Disable the MCP server",
		Long:  `Disable the MCP server by removing it to claude's MCP config file`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO
			return nil
		},
	}

	return cmd
}
