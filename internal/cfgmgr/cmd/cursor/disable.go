package cursor

import (
	"github.com/njayp/ophis/internal/cfgmgr/manager"
	"github.com/spf13/cobra"
)

type disableFlags struct {
	defaultServerName string
	configPath        string
	serverName        string
	workspace         bool
}

// disableCommand creates a Cobra command for removing an MCP server from Cursor.
// serverName is the default MCP server entry name; the --server-name flag overrides it.
func disableCommand(serverName string) *cobra.Command {
	f := &disableFlags{defaultServerName: serverName}
	cmd := &cobra.Command{
		Use:   "disable",
		Short: "Remove server from Cursor config",
		Long:  "Remove this application from Cursor MCP servers",
		RunE: func(_ *cobra.Command, _ []string) error {
			return f.run()
		},
	}

	// Add flags
	flags := cmd.Flags()
	flags.StringVar(&f.configPath, "config-path", "", "Path to Cursor config file")
	flags.StringVar(&f.serverName, "server-name", "", manager.ServerNameRemoveUsage(serverName))
	flags.BoolVar(&f.workspace, "workspace", false, "Remove from workspace settings (.cursor/mcp.json) instead of user settings")

	return cmd
}

func (f *disableFlags) run() error {
	// Resolve the server name (flag → configured default → executable name).
	serverName, err := manager.ResolveServerName(f.serverName, f.defaultServerName)
	if err != nil {
		return err
	}

	m, err := manager.NewCursorManager(f.configPath, f.workspace)
	if err != nil {
		return err
	}

	return m.DisableServer(serverName)
}
