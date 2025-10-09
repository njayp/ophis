package cursor

import (
	"fmt"

	"github.com/njayp/ophis/internal/cfgmgr"
	"github.com/njayp/ophis/internal/cfgmgr/cursor/config"
	"github.com/spf13/cobra"
)

type listCommandFlags struct {
	configPath string
	workspace  bool
}

// listCommand creates a Cobra command for listing MCP servers in Cursor.
func listCommand() *cobra.Command {
	listFlags := &listCommandFlags{}
	cmd := &cobra.Command{
		Use:   cfgmgr.CmdList,
		Short: cmdListShort,
		Long:  cmdListLong,
		RunE: func(_ *cobra.Command, _ []string) error {
			return listMCPServers(listFlags)
		},
	}

	// Add flags
	flags := cmd.Flags()
	flags.StringVar(&listFlags.configPath, cfgmgr.FlagConfigPath, "", "Path to Cursor config file")
	flags.BoolVar(&listFlags.workspace, cfgmgr.FlagWorkspace, false, "List from workspace settings (.cursor/mcp.json) instead of user settings")

	return cmd
}

func listMCPServers(flags *listCommandFlags) error {
	// Determine configuration type
	configType := config.UserConfig
	if flags.workspace {
		configType = config.WorkspaceConfig
	}

	// Create config manager
	configManager := config.NewCursorConfigManager(flags.configPath, configType)

	// Load config
	cursorConfig, err := configManager.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load Cursor configuration: %w", err)
	}

	fmt.Printf("Cursor MCP Servers (%s configuration):\n", configType)
	fmt.Printf("Configuration file: %s\n\n", configManager.GetConfigPath())

	if len(cursorConfig.MCPServers) == 0 {
		fmt.Println(cfgmgr.MsgNoServersConfigured)
		return nil
	}

	for name, server := range cursorConfig.MCPServers {
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
