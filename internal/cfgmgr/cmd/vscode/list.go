package vscode

import (
	"fmt"

	"github.com/njayp/ophis/internal/cfgmgr/manager"
	"github.com/njayp/ophis/internal/cfgmgr/manager/vscode"
	"github.com/spf13/cobra"
)

type listCommandFlags struct {
	configPath string
	workspace  bool
}

// listCommand creates a Cobra command for listing MCP servers in VSCode.
func listCommand() *cobra.Command {
	listFlags := &listCommandFlags{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Show VSCode MCP servers",
		Long:  "Show all MCP servers configured in VSCode",
		RunE: func(_ *cobra.Command, _ []string) error {
			return listFlags.listMCPServers()
		},
	}

	// Add flags
	flags := cmd.Flags()
	flags.StringVar(&listFlags.configPath, "config-path", "", "Path to VSCode config file")
	flags.BoolVar(&listFlags.workspace, "workspace", false, "List from workspace settings (.vscode/mcp.json) instead of user settings")

	return cmd
}

func (f *listCommandFlags) listMCPServers() error {
	// Create config manager
	manager := manager.Manager[vscode.Config, vscode.MCPServer]{
		Platform: vscode.NewVSCodeConfigManager(f.workspace),
	}

	config := vscode.Config{}
	err := manager.LoadJSONConfig(&config)
	if err != nil {
		return err
	}

	if len(config.Servers) == 0 {
		fmt.Println("")
		return nil
	}

	for name, server := range config.Servers {
		fmt.Printf("Server: %s\n", name)
		fmt.Printf("  Type: %s\n", server.Type)
		if server.Command != "" {
			fmt.Printf("  Command: %s\n", server.Command)
		}
		if server.URL != "" {
			fmt.Printf("  URL: %s\n", server.URL)
		}
		if len(server.Args) > 0 {
			fmt.Printf("  Args: %v\n", server.Args)
		}
		if len(server.Env) > 0 {
			fmt.Printf("  Environment:\n")
			for key, value := range server.Env {
				fmt.Printf("    %s: %s\n", key, value)
			}
		}
		if len(server.Headers) > 0 {
			fmt.Printf("  Headers:\n")
			for key, value := range server.Headers {
				fmt.Printf("    %s: %s\n", key, value)
			}
		}
		fmt.Println()
	}

	return nil
}
