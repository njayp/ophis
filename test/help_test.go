package test

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/njayp/ophis"
)

// updateGolden regenerates the golden files instead of asserting against them.
// Run: go test ./test -run 'TestUsage|TestHelp' -update
var updateGolden = flag.Bool("update", false, "update help/usage golden files")

// TestUsage asserts the `Usage()` output of every command in the ophis tree
// against a golden file, so changes to flag help strings are caught in review.
func TestUsage(t *testing.T) {
	cmd := ophis.Command(nil)
	walkCommandOutput(t, cmd.Name(), cmd, "usage", func(c *cobra.Command) error {
		return c.Usage()
	})
}

// TestHelp asserts the `Help()` output of every command in the ophis tree
// against a golden file.
func TestHelp(t *testing.T) {
	cmd := ophis.Command(nil)
	walkCommandOutput(t, cmd.Name(), cmd, "help", func(c *cobra.Command) error {
		return c.Help()
	})
}

// TestEnableServerNameUsageFromConfig checks that a configured Config.ServerName
// is reflected in the `--server-name` flag help on every editor's `enable`
// command. The golden tests above cover the empty-ServerName branch (nil config);
// this covers the configured branch.
func TestEnableServerNameUsageFromConfig(t *testing.T) {
	cmd := ophis.Command(&ophis.Config{ServerName: "my-cli"})
	for _, editor := range []string{"claude", "vscode", "cursor"} {
		t.Run(editor, func(t *testing.T) {
			enable := findCommand(t, cmd, editor, "enable")
			flag := enable.Flags().Lookup("server-name")
			require.NotNil(t, flag, "enable should have a --server-name flag")
			assert.Equal(t, `Name for the MCP server (default: "my-cli")`, flag.Usage)
		})
	}
}

// findCommand walks from parent down the given subcommand path, failing the test
// if any segment is missing.
func findCommand(t *testing.T, parent *cobra.Command, path ...string) *cobra.Command {
	t.Helper()
	cmd := parent
	for _, name := range path {
		var next *cobra.Command
		for _, sub := range cmd.Commands() {
			if sub.Name() == name {
				next = sub
				break
			}
		}
		require.NotNil(t, next, "command %q not found under %q", name, cmd.Name())
		cmd = next
	}
	return cmd
}

// walkCommandOutput renders cmd (and each of its subcommands, recursively) with
// render, comparing the captured output to testdata/<baseDir>/<path>.txt. prefix
// is the command path used to name the golden file (e.g. "mcp/claude/enable").
func walkCommandOutput(t *testing.T, prefix string, cmd *cobra.Command, baseDir string, render func(*cobra.Command) error) {
	t.Helper()
	t.Run(cmd.Name(), func(t *testing.T) {
		buf := &bytes.Buffer{}
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		require.NoError(t, render(cmd))

		assertGolden(t, buf.String(), filepath.Join(baseDir, prefix+".txt"))

		for _, sub := range cmd.Commands() {
			walkCommandOutput(t, filepath.Join(prefix, sub.Name()), sub, baseDir, render)
		}
	})
}

// assertGolden compares got to the golden file at testdata/<name>, or rewrites
// the golden file when -update is set.
func assertGolden(t *testing.T, got, name string) {
	t.Helper()
	goldenPath := filepath.Join("testdata", name)

	if *updateGolden {
		require.NoError(t, os.MkdirAll(filepath.Dir(goldenPath), 0o755))
		require.NoError(t, os.WriteFile(goldenPath, []byte(got), 0o644))
		return
	}

	want, err := os.ReadFile(goldenPath)
	require.NoError(t, err, "missing golden file %s (run: go test ./test -update)", goldenPath)
	assert.Equal(t, string(want), got, "help output for %s changed (run: go test ./test -update to accept)", name)
}
