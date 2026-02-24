package zed

import (
	"fmt"

	"github.com/njayp/ophis/internal/cfgmgr/manager"
	"github.com/spf13/cobra"
)

type listFlags struct {
	configPath string
	workspace  bool
}

// listCommand creates a Cobra command for listing configured MCP servers in Zed.
func listCommand() *cobra.Command {
	f := &listFlags{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Show Zed MCP context servers",
		Long:  "Show all MCP context servers configured in Zed",
		RunE: func(_ *cobra.Command, _ []string) error {
			return f.run()
		},
	}

	// Add flags
	flags := cmd.Flags()
	flags.StringVar(&f.configPath, "config-path", "", "Path to Zed settings file")
	flags.BoolVar(&f.workspace, "workspace", false, "List from workspace settings (.zed/settings.json) instead of user settings")

	return cmd
}

func (f *listFlags) run() error {
	m, err := manager.NewZedManager(f.configPath, f.workspace)
	if err != nil {
		return err
	}

	fmt.Printf("Zed MCP context servers:\n\n")
	m.Print()
	return nil
}
