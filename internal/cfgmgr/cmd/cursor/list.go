package cursor

import (
	"fmt"

	"github.com/njayp/ophis/internal/cfgmgr/manager"
	"github.com/spf13/cobra"
)

type listFlags struct {
	configPath string
	workspace  bool
}

// listCommand creates a Cobra command for listing configured MCP servers in Cursor.
func listCommand() *cobra.Command {
	f := &listFlags{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Show Cursor MCP servers",
		Long:  "Show all MCP servers configured in Cursor",
		RunE: func(_ *cobra.Command, _ []string) error {
			return f.run()
		},
	}

	// Add flags
	flags := cmd.Flags()
	flags.StringVar(&f.configPath, "config-path", "", "Path to Cursor config file")
	flags.BoolVar(&f.workspace, "workspace", false, "List from workspace settings (.cursor/mcp.json) instead of user settings")

	return cmd
}

func (f *listFlags) run() error {
	m, err := manager.NewCursorManager(f.configPath, f.workspace)
	if err != nil {
		return err
	}

	fmt.Printf("Cursor MCP servers:\n\n")
	m.Print()
	return nil
}
