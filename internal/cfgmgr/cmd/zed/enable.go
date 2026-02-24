package zed

import (
	"fmt"
	"os"

	"github.com/njayp/ophis/internal/cfgmgr/manager"
	"github.com/njayp/ophis/internal/cfgmgr/manager/zed"
	"github.com/spf13/cobra"
)

type enableFlags struct {
	commandName string
	defaultEnv  map[string]string
	configPath  string
	logLevel    string
	serverName  string
	workspace   bool
	env         map[string]string
}

// enableCommand creates a Cobra command for adding an MCP server to Zed.
// commandName is the Use name of the ophis root command (e.g. "mcp" or "agent").
// defaultEnv is merged into the server env; user-provided --env values take precedence.
func enableCommand(commandName string, defaultEnv map[string]string) *cobra.Command {
	f := &enableFlags{commandName: commandName, defaultEnv: defaultEnv}
	cmd := &cobra.Command{
		Use:   "enable",
		Short: "Add server to Zed config",
		Long:  "Add this application as an MCP context server in Zed",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return f.run(cmd)
		},
	}

	// Add flags
	flags := cmd.Flags()
	flags.StringVar(&f.logLevel, "log-level", "", "Log level (debug, info, warn, error)")
	flags.StringVar(&f.configPath, "config-path", "", "Path to Zed settings file")
	flags.StringVar(&f.serverName, "server-name", "", "Name for the MCP server (default: derived from executable name)")
	flags.BoolVar(&f.workspace, "workspace", false, "Add to workspace settings (.zed/settings.json) instead of user settings")
	flags.StringToStringVarP(&f.env, "env", "e", nil, "Environment variables (e.g., --env KEY1=value1 --env KEY2=value2)")

	return cmd
}

func (f *enableFlags) run(cmd *cobra.Command) error {
	// Get the current executable path
	executablePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to determine executable path: %w", err)
	}

	// Build server configuration
	mcpPath, err := manager.GetCmdPath(cmd, f.commandName)
	if err != nil {
		return fmt.Errorf("failed to determine MCP command path: %w", err)
	}

	server := zed.Server{
		Command: executablePath,
		Args:    append(mcpPath, "start"),
	}

	// Add log level to args if specified
	if f.logLevel != "" {
		server.Args = append(server.Args, "--log-level", f.logLevel)
	}

	// Merge default env with user-provided env.
	// User values take precedence on conflict.
	env := make(map[string]string)
	for k, v := range f.defaultEnv {
		env[k] = v
	}
	for k, v := range f.env {
		env[k] = v
	}
	if len(env) > 0 {
		server.Env = env
	}

	if f.serverName == "" {
		f.serverName = manager.DeriveServerName(executablePath)
	}

	m, err := manager.NewZedManager(f.configPath, f.workspace)
	if err != nil {
		return err
	}

	return m.EnableServer(f.serverName, server)
}
