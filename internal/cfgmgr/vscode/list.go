package vscode

import (
	"fmt"

	"github.com/njayp/ophis/internal/cfgmgr/vscode/config"
	"github.com/spf13/cobra"
)

type listCommandFlags struct {
	configPath string
	workspace  bool
	configType string
}

// listCommand creates a Cobra command for listing MCP servers in VSCode.
func listCommand() *cobra.Command {
	listFlags := &listCommandFlags{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Show VSCode MCP servers",
		Long:  `Show all MCP servers configured in VSCode`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return listMCPServers(listFlags)
		},
	}

	// Add flags
	flags := cmd.Flags()
	flags.StringVar(&listFlags.configPath, "config-path", "", "Path to VSCode config file")
	flags.BoolVar(&listFlags.workspace, "workspace", false, "List from workspace settings (.vscode/mcp.json) instead of user settings")
	flags.StringVar(&listFlags.configType, "config-type", "", "Configuration type: 'workspace' or 'user' (default: user)")

	return cmd
}

func listMCPServers(flags *listCommandFlags) error {
	// Determine configuration type
	configType := config.UserConfig
	if flags.workspace || flags.configType == "workspace" {
		configType = config.WorkspaceConfig
	} else if flags.configType == "user" {
		configType = config.UserConfig
	} else if flags.configType != "" {
		return fmt.Errorf("invalid config type '%s': must be 'workspace' or 'user'", flags.configType)
	}

	// Create config manager
	configManager := config.NewVSCodeConfigManager(flags.configPath, configType)

	// Load config
	vsConfig, err := configManager.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load VSCode configuration: %w", err)
	}

	configTypeStr := "user"
	if configType == config.WorkspaceConfig {
		configTypeStr = "workspace"
	}

	fmt.Printf("VSCode MCP Servers (%s configuration):\n", configTypeStr)
	fmt.Printf("Configuration file: %s\n\n", configManager.GetConfigPath())

	if len(vsConfig.Servers) == 0 {
		fmt.Printf("No MCP servers configured.\n")
		return nil
	}

	for name, server := range vsConfig.Servers {
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
