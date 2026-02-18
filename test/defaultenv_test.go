package test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/njayp/ophis"
	"github.com/njayp/ophis/internal/cfgmgr/manager/claude"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createDefaultEnvCommand creates a command tree with DefaultEnv configured.
// configPath is passed to the enable command via --config-path so the test
// doesn't touch the real Claude Desktop config.
func createDefaultEnvCommand(defaultEnv map[string]string) *cobra.Command {
	root := &cobra.Command{
		Use:   "testcli",
		Short: "Test CLI for DefaultEnv",
	}

	status := &cobra.Command{
		Use:   "status",
		Short: "Show status",
		Run:   func(_ *cobra.Command, _ []string) {},
	}

	root.AddCommand(status)
	root.AddCommand(ophis.Command(&ophis.Config{
		CommandName: "agent",
		DefaultEnv:  defaultEnv,
	}))

	return root
}

// readClaudeConfig reads and parses a Claude Desktop config file, returning
// the server entry with the given name.
func readClaudeConfig(t *testing.T, configPath, serverName string) claude.Server {
	t.Helper()
	data, err := os.ReadFile(configPath)
	require.NoError(t, err, "failed to read config file")

	var cfg claude.Config
	err = json.Unmarshal(data, &cfg)
	require.NoError(t, err, "failed to parse config file")

	require.True(t, cfg.HasServer(serverName), "server %q not found in config", serverName)
	return cfg.Servers[serverName]
}

func TestDefaultEnvWrittenToConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "claude_config.json")

	cmd := createDefaultEnvCommand(map[string]string{
		"PATH":       "/usr/local/bin:/usr/bin",
		"KUBECONFIG": "/home/user/.kube/config",
	})

	cmd.SetArgs([]string{"agent", "claude", "enable", "--config-path", configPath, "--server-name", "testcli"})
	err := cmd.Execute()
	require.NoError(t, err)

	server := readClaudeConfig(t, configPath, "testcli")
	require.NotNil(t, server.Env, "server env should not be nil")
	assert.Equal(t, "/usr/local/bin:/usr/bin", server.Env["PATH"])
	assert.Equal(t, "/home/user/.kube/config", server.Env["KUBECONFIG"])
}

func TestDefaultEnvUserOverride(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "claude_config.json")

	cmd := createDefaultEnvCommand(map[string]string{
		"PATH": "/default/path",
		"HOME": "/default/home",
	})

	cmd.SetArgs([]string{
		"agent", "claude", "enable",
		"--config-path", configPath,
		"--server-name", "testcli",
		"--env", "PATH=/user/path",
		"--env", "EXTRA=value",
	})
	err := cmd.Execute()
	require.NoError(t, err)

	server := readClaudeConfig(t, configPath, "testcli")
	require.NotNil(t, server.Env, "server env should not be nil")

	// User value overrides default.
	assert.Equal(t, "/user/path", server.Env["PATH"], "user --env should override DefaultEnv")
	// Default value preserved when no user override.
	assert.Equal(t, "/default/home", server.Env["HOME"], "unoverridden DefaultEnv should be preserved")
	// User-only value is present.
	assert.Equal(t, "value", server.Env["EXTRA"], "user-only --env should be present")
}

func TestNilDefaultEnvNoEnvBlock(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "claude_config.json")

	// nil DefaultEnv, no --env flag — should produce no env in config.
	cmd := createDefaultEnvCommand(nil)
	cmd.SetArgs([]string{"agent", "claude", "enable", "--config-path", configPath, "--server-name", "testcli"})
	err := cmd.Execute()
	require.NoError(t, err)

	server := readClaudeConfig(t, configPath, "testcli")
	assert.Empty(t, server.Env, "no env should be written when DefaultEnv is nil and no --env given")
}

func TestNilDefaultEnvWithUserEnv(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "claude_config.json")

	cmd := createDefaultEnvCommand(nil)
	cmd.SetArgs([]string{
		"agent", "claude", "enable",
		"--config-path", configPath,
		"--server-name", "testcli",
		"--env", "FOO=bar",
	})
	err := cmd.Execute()
	require.NoError(t, err)

	server := readClaudeConfig(t, configPath, "testcli")
	require.NotNil(t, server.Env)
	assert.Equal(t, "bar", server.Env["FOO"])
	assert.Len(t, server.Env, 1, "only user-provided env should be present")
}

func TestEmptyDefaultEnvNoEnvBlock(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "claude_config.json")

	// Empty map (not nil) — same behavior as nil: no env written.
	cmd := createDefaultEnvCommand(map[string]string{})
	cmd.SetArgs([]string{"agent", "claude", "enable", "--config-path", configPath, "--server-name", "testcli"})
	err := cmd.Execute()
	require.NoError(t, err)

	server := readClaudeConfig(t, configPath, "testcli")
	assert.Empty(t, server.Env, "no env should be written when DefaultEnv is empty and no --env given")
}
