package claude

import (
	"fmt"

	"github.com/njayp/ophis/internal/cfgmgr/manager"
	"github.com/njayp/ophis/internal/cfgmgr/manager/claude"
	"github.com/spf13/cobra"
)

type listCommandFlags struct {
	configPath string
}

// listCommand creates a Cobra command for listing configured MCP servers.
func listCommand() *cobra.Command {
	listFlags := &listCommandFlags{} // Reuse flags struct for config-path
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Show Claude MCP servers",
		Long:  "Show all MCP servers configured in Claude Desktop",
		RunE: func(_ *cobra.Command, _ []string) error {
			return listFlags.listMCPServers()
		},
	}

	// Add flags
	flags := cmd.Flags()
	flags.StringVar(&listFlags.configPath, "config-path", "", "Path to Claude config file")
	return cmd
}

func (f *listCommandFlags) listMCPServers() error {
	// Create config manager
	manager := manager.Manager[claude.Config, claude.MCPServer]{
		Platform: claude.NewClaudeCodeConfigManager(),
	}

	config := claude.Config{}
	err := manager.LoadJSONConfig(&config)
	if err != nil {
		return err
	}

	if len(config.MCPServers) == 0 {
		fmt.Println("")
		return nil
	}

	for name, server := range config.MCPServers {
		fmt.Printf("Server: %s\n", name)
		fmt.Printf("  Command: %s\n", server.Command)
		if len(server.Args) > 0 {
			fmt.Printf("  Args: %v\n", server.Args)
		}
		if len(server.Env) > 0 {
			fmt.Printf("  Environment:\n")
			for key, value := range server.Env {
				fmt.Printf("    %s: %s\n", key, value)
			}
		}
		fmt.Println()
	}

	return nil
}
