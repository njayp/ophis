package claude

import (
	"fmt"
	"os"

	"github.com/njayp/ophis/mcp/claude/config"
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
		Short: "List configured MCP servers",
		Long:  `List all MCP servers currently configured in Claude's MCP config file`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return listMCPServers(listFlags)
		},
	}

	// Add flags
	flags := cmd.Flags()
	flags.StringVar(&listFlags.configPath, "config-path", "", "Path to Claude config file")
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
		fmt.Println("No MCP servers are currently configured.")
		fmt.Println("\nTo enable this application as an MCP server, run:")
		fmt.Println("  <your-app> mcp enable")
		return nil
	}

	fmt.Printf("Configured MCP servers (%d):\n\n", len(claudeConfig.MCPServers))
	for name, server := range claudeConfig.MCPServers {
		fmt.Printf("  📦 %s\n", name)
		fmt.Printf("     Command: %s\n", server.Command)
		if len(server.Args) > 0 {
			fmt.Printf("     Args: %v\n", server.Args)
		}
		if len(server.Env) > 0 {
			fmt.Printf("     Environment: %v\n", server.Env)
		}

		// Check if the executable still exists
		if _, err := os.Stat(server.Command); os.IsNotExist(err) {
			fmt.Printf("     ⚠️  Warning: Executable not found\n")
		}
		fmt.Println()
	}

	fmt.Println("💡 Remember to restart Claude Desktop after making changes to the configuration.")
	return nil
}
