package claude

import (
	"github.com/njayp/ophis/internal/cfgmgr/manager"
	"github.com/spf13/cobra"
)

type disableFlags struct {
	defaultServerName string
	configPath        string
	serverName        string
}

// disableCommand creates a Cobra command for removing an MCP server from Claude Desktop.
// serverName is the default MCP server entry name; the --server-name flag overrides it.
func disableCommand(serverName string) *cobra.Command {
	f := &disableFlags{defaultServerName: serverName}
	cmd := &cobra.Command{
		Use:   "disable",
		Short: "Remove server from Claude config",
		Long:  "Remove this application from Claude Desktop MCP servers",
		RunE: func(_ *cobra.Command, _ []string) error {
			return f.run()
		},
	}

	// Add flags
	flags := cmd.Flags()
	flags.StringVar(&f.configPath, "config-path", "", "Path to Claude config file")
	flags.StringVar(&f.serverName, "server-name", "", manager.ServerNameRemoveUsage(serverName))
	return cmd
}

func (f *disableFlags) run() error {
	// Resolve the server name (flag → configured default → executable name).
	serverName, err := manager.ResolveServerName(f.serverName, f.defaultServerName)
	if err != nil {
		return err
	}

	m, err := manager.NewClaudeManager(f.configPath)
	if err != nil {
		return err
	}

	return m.DisableServer(serverName)
}
