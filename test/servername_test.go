package test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/njayp/ophis"
	"github.com/njayp/ophis/internal/cfgmgr/manager"
	"github.com/njayp/ophis/internal/cfgmgr/manager/claude"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createServerNameCommand creates a command tree with Config.ServerName set.
// The enable command is invoked with --config-path so the test never touches
// the real Claude Desktop config.
func createServerNameCommand(serverName string) *cobra.Command {
	root := &cobra.Command{
		Use:   "testcli",
		Short: "Test CLI for ServerName",
	}

	status := &cobra.Command{
		Use:   "status",
		Short: "Show status",
		Run:   func(_ *cobra.Command, _ []string) {},
	}

	root.AddCommand(status)
	root.AddCommand(ophis.Command(&ophis.Config{
		CommandName: "agent",
		ServerName:  serverName,
	}))

	return root
}

// readClaudeConfigNames reads a Claude Desktop config file and returns the
// names of all configured servers.
func readClaudeConfigNames(t *testing.T, configPath string) []string {
	t.Helper()
	data, err := os.ReadFile(configPath)
	require.NoError(t, err, "failed to read config file")

	var cfg claude.Config
	err = json.Unmarshal(data, &cfg)
	require.NoError(t, err, "failed to parse config file")

	names := make([]string, 0, len(cfg.Servers))
	for name := range cfg.Servers {
		names = append(names, name)
	}
	return names
}

func TestServerNameFromConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "claude_config.json")

	cmd := createServerNameCommand("configured-name")
	cmd.SetArgs([]string{"agent", "claude", "enable", "--config-path", configPath})
	err := cmd.Execute()
	require.NoError(t, err)

	names := readClaudeConfigNames(t, configPath)
	assert.Equal(t, []string{"configured-name"}, names, "server should use the configured ServerName")
}

func TestServerNameFlagOverridesConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "claude_config.json")

	cmd := createServerNameCommand("configured-name")
	cmd.SetArgs([]string{
		"agent", "claude", "enable",
		"--config-path", configPath,
		"--server-name", "flag-name",
	})
	err := cmd.Execute()
	require.NoError(t, err)

	names := readClaudeConfigNames(t, configPath)
	assert.Equal(t, []string{"flag-name"}, names, "--server-name should override the configured ServerName")
}

func TestServerNameEnableDisableRoundTrip(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "claude_config.json")

	// enable writes an entry named after the configured ServerName.
	enable := createServerNameCommand("configured-name")
	enable.SetArgs([]string{"agent", "claude", "enable", "--config-path", configPath})
	require.NoError(t, enable.Execute())

	names := readClaudeConfigNames(t, configPath)
	require.Equal(t, []string{"configured-name"}, names, "server should use the configured ServerName")

	// disable (without --server-name) must resolve the same configured name
	// and remove the entry, leaving nothing behind.
	disable := createServerNameCommand("configured-name")
	disable.SetArgs([]string{"agent", "claude", "disable", "--config-path", configPath})
	require.NoError(t, disable.Execute())

	names = readClaudeConfigNames(t, configPath)
	assert.Empty(t, names, "disable should remove the entry created by enable")
}

func TestServerNameEmptyFallsBackToExecutable(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "claude_config.json")

	// Empty ServerName and no --server-name flag: fall back to the
	// executable-derived name.
	cmd := createServerNameCommand("")
	cmd.SetArgs([]string{"agent", "claude", "enable", "--config-path", configPath})
	err := cmd.Execute()
	require.NoError(t, err)

	executablePath, err := os.Executable()
	require.NoError(t, err)
	expected := manager.DeriveServerName(executablePath)

	names := readClaudeConfigNames(t, configPath)
	assert.Equal(t, []string{expected}, names, "empty ServerName should fall back to the executable name")
}
