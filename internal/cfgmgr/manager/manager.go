package manager

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/njayp/ophis/internal/cfgmgr/manager/claude"
	"github.com/njayp/ophis/internal/cfgmgr/manager/cursor"
	"github.com/njayp/ophis/internal/cfgmgr/manager/vscode"
	"github.com/njayp/ophis/internal/cfgmgr/manager/zed"
)

// Config represents MCP server configuration that can be managed.
type Config[S Server] interface {
	HasServer(name string) bool
	AddServer(name string, server S)
	RemoveServer(name string)
	Print()
}

// Server represents an individual MCP server entry.
type Server interface {
	Print()
}

// Preprocessor is an optional interface that a Config may implement to
// transform raw file bytes before JSON unmarshaling. This is used by
// configs whose backing file is not strict JSON (e.g. Zed's settings.json
// is JSONC — JSON with comments and trailing commas).
type Preprocessor interface {
	Preprocess(data []byte) ([]byte, error)
}

// Manager provides generic configuration management for MCP servers.
// It handles loading, saving, and modifying MCP server configurations.
// It is not thread-safe.
type Manager[S Server, C Config[S]] struct {
	configPath string
	config     C
}

// NewVSCodeManager creates a new Manager configured for VSCode MCP servers.
// If workspace is true, uses workspace configuration (.vscode/mcp.json),
// otherwise uses user-level configuration.
func NewVSCodeManager(configPath string, workspace bool) (*Manager[vscode.Server, *vscode.Config], error) {
	if configPath == "" {
		configPath = vscode.ConfigPath(workspace)
	}

	m := &Manager[vscode.Server, *vscode.Config]{
		config:     &vscode.Config{},
		configPath: configPath,
	}

	return m, m.loadConfig()
}

// NewCursorManager creates a new Manager configured for Cursor MCP servers.
// If workspace is true, uses workspace configuration (.cursor/mcp.json),
// otherwise uses user-level configuration.
func NewCursorManager(configPath string, workspace bool) (*Manager[cursor.Server, *cursor.Config], error) {
	if configPath == "" {
		configPath = cursor.ConfigPath(workspace)
	}

	m := &Manager[cursor.Server, *cursor.Config]{
		config:     &cursor.Config{},
		configPath: configPath,
	}

	return m, m.loadConfig()
}

// NewZedManager creates a new Manager configured for Zed MCP context servers.
// If workspace is true, uses workspace configuration (.zed/settings.json),
// otherwise uses user-level configuration (~/.config/zed/settings.json).
func NewZedManager(configPath string, workspace bool) (*Manager[zed.Server, *zed.Config], error) {
	if configPath == "" {
		configPath = zed.ConfigPath(workspace)
	}

	m := &Manager[zed.Server, *zed.Config]{
		config:     &zed.Config{},
		configPath: configPath,
	}

	return m, m.loadConfig()
}

// NewClaudeManager creates a new Manager configured for Claude Desktop MCP servers.
func NewClaudeManager(configPath string) (*Manager[claude.Server, *claude.Config], error) {
	if configPath == "" {
		configPath = claude.ConfigPath()
	}

	m := &Manager[claude.Server, *claude.Config]{
		config:     &claude.Config{},
		configPath: configPath,
	}

	return m, m.loadConfig()
}

// EnableServer adds or updates an MCP server in the configuration.
func (m *Manager[S, C]) EnableServer(name string, server S) error {
	if m.config.HasServer(name) {
		fmt.Printf("⚠️  MCP server %q already exists and will be overwritten\n", name)
	}

	m.config.AddServer(name, server)
	err := m.saveConfig()
	if err != nil {
		return err
	}

	fmt.Printf("Successfully enabled MCP server: %q\n", name)
	server.Print()
	return nil
}

// DisableServer removes an MCP server from the configuration.
func (m *Manager[S, C]) DisableServer(name string) error {
	if m.config.HasServer(name) {
		m.config.RemoveServer(name)
		return m.saveConfig()
	}

	fmt.Printf("⚠️  MCP server %q does not exist\n", name)
	return nil
}

// loadConfig unmarshals a JSON file into the provided interface.
func (m *Manager[S, C]) loadConfig() error {
	fmt.Printf("Using config path %q\n", m.configPath)

	// Check if config file exists
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		// File doesn't exist - return nil to allow initialization
		return nil
	}

	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return fmt.Errorf("failed to read configuration file at %q: %w", m.configPath, err)
	}

	// If the config knows how to preprocess its raw bytes (e.g. strip JSONC
	// comments and trailing commas), do that before standard JSON parsing.
	if pp, ok := any(m.config).(Preprocessor); ok {
		data, err = pp.Preprocess(data)
		if err != nil {
			return fmt.Errorf("failed to preprocess configuration file at %q: %w", m.configPath, err)
		}
	}

	if err := json.Unmarshal(data, m.config); err != nil {
		return fmt.Errorf("failed to parse configuration file at %q: invalid JSON format: %w", m.configPath, err)
	}

	return nil
}

// saveConfig marshals and saves configuration as formatted JSON.
func (m *Manager[S, C]) saveConfig() error {
	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(m.configPath), 0o755); err != nil {
		return fmt.Errorf("failed to create directory for configuration file at %q: %w", filepath.Dir(m.configPath), err)
	}

	data, err := json.MarshalIndent(m.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal configuration to JSON: %w", err)
	}

	// backup file
	err = m.backupConfig()
	if err != nil {
		return err
	}

	if err := os.WriteFile(m.configPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write configuration file at %q: %w", m.configPath, err)
	}

	return nil
}

// Print calls Print on the underlying config
func (m *Manager[S, C]) Print() {
	m.config.Print()
}

func (m *Manager[S, C]) backupConfig() error {
	// Check if config file exists
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		// File doesn't exist - nothing to backup
		return nil
	}

	sourceFile, err := os.Open(m.configPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = sourceFile.Close()
	}()

	ext := filepath.Ext(m.configPath)
	dest := strings.TrimSuffix(m.configPath, ext) + ".backup.json"
	fmt.Printf("Backing up config file at %q\n", dest)
	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer func() {
		_ = destFile.Close()
	}()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
