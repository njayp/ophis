package claude

import (
	"fmt"

	"github.com/njayp/ophis/internal/cfgmgr/manager"
	"github.com/spf13/cobra"
)

type listFlags struct {
	configPath string
}

// listCommand creates a Cobra command for listing configured MCP servers in Claude Desktop.
func listCommand() *cobra.Command {
	f := &listFlags{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Show Claude MCP servers",
		Long:  "Show all MCP servers configured in Claude Desktop",
		RunE: func(_ *cobra.Command, _ []string) error {
			return f.run()
		},
	}

	// Add flags
	flags := cmd.Flags()
	flags.StringVar(&f.configPath, "config-path", "", "Path to Claude config file")
	return cmd
}

func (f *listFlags) run() error {
	m, err := manager.NewClaudeManager(f.configPath)
	if err != nil {
		return err
	}

	fmt.Printf("Claude Desktop MCP servers:\n\n")
	m.Print()
	return nil
}
