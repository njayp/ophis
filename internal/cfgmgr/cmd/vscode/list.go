package vscode

import (
	"fmt"

	"github.com/njayp/ophis/internal/cfgmgr/manager"
	"github.com/spf13/cobra"
)

type listFlags struct {
	configPath string
	workspace  bool
}

// listCommand creates a Cobra command for listing configured MCP servers in VSCode.
func listCommand() *cobra.Command {
	f := &listFlags{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Show VSCode MCP servers",
		Long:  "Show all MCP servers configured in VSCode",
		RunE: func(_ *cobra.Command, _ []string) error {
			return f.run()
		},
	}

	// Add flags
	flags := cmd.Flags()
	flags.StringVar(&f.configPath, "config-path", "", "Path to VSCode config file")
	flags.BoolVar(&f.workspace, "workspace", false, "List from workspace settings (.vscode/mcp.json) instead of user settings")

	return cmd
}

func (f *listFlags) run() error {
	m, err := manager.NewVSCodeManager(f.configPath, f.workspace)
	if err != nil {
		return err
	}

	fmt.Printf("VSCode MCP servers:\n\n")
	m.Print()
	return nil
}
