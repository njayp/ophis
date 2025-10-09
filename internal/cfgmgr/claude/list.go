package claude

import (
	"fmt"

	"github.com/njayp/ophis/internal/cfgmgr"
	"github.com/njayp/ophis/internal/cfgmgr/claude/config"
	"github.com/spf13/cobra"
)

type listCommandFlags struct {
	configPath string
}

// listCommand creates a Cobra command for listing configured MCP servers.
func listCommand() *cobra.Command {
	listFlags := &listCommandFlags{} // Reuse flags struct for config-path
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
	flags.StringVar(&listFlags.configPath, cfgmgr.FlagConfigPath, "", "Path to Claude config file")
	return cmd
}

func listMCPServers(flags *listCommandFlags) error {
	// Create config manager
	configManager := config.NewClaudeConfigManager(flags.configPath)

	// Load config
	claudeConfig, err := configManager.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load Claude configuration: %w", err)
	}

	fmt.Printf("Claude MCP Configuration File: %s\n\n", configManager.GetConfigPath())

	if len(claudeConfig.MCPServers) == 0 {
		fmt.Println(cfgmgr.MsgNoServersConfigured)
		fmt.Println("\nTo enable this application as an MCP server, run:")
		fmt.Println("  <your-app> mcp claude enable")
		return nil
	}

	fmt.Printf("Configured MCP servers (%d):\n\n", len(claudeConfig.MCPServers))
	for name, server := range claudeConfig.MCPServers {
		fmt.Printf("Server: %s\n", name)
		fmt.Printf("  Command: %s\n", server.Command)
		if len(server.Args) > 0 {
			fmt.Printf("  Args: %v\n", server.Args)
		}
		if len(server.Env) > 0 {
			fmt.Printf("  Environment: %v\n", server.Env)
		}

		fmt.Println()
	}

	fmt.Println("💡 Remember to restart Claude Desktop after making changes to the configuration.")
	return nil
}
