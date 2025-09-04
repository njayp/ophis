package cfgmgr

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestBackupConfigFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.json")

	// Test backing up a non-existent file (should return nil)
	err := BackupConfigFile(configPath)
	if err != nil {
		t.Errorf("BackupConfigFile should return nil for non-existent file, got: %v", err)
	}

	// Create a test config file
	testData := []byte(`{"test": "data"}`)
	if err := os.WriteFile(configPath, testData, 0o644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Test backing up an existing file
	err = BackupConfigFile(configPath)
	if err != nil {
		t.Errorf("BackupConfigFile failed: %v", err)
	}

	// Verify backup was created
	backupPath := configPath + ".backup"
	backupData, err := os.ReadFile(backupPath)
	if err != nil {
		t.Errorf("Failed to read backup file: %v", err)
	}

	if string(backupData) != string(testData) {
		t.Errorf("Backup content mismatch. Expected %s, got %s", testData, backupData)
	}
}

func TestLoadJSONConfig(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.json")

	type TestConfig struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	// Test loading from non-existent file (should not error)
	var config TestConfig
	err := LoadJSONConfig(configPath, &config)
	if err != nil {
		t.Errorf("LoadJSONConfig should return nil for non-existent file, got: %v", err)
	}

	// Create a test config file
	testConfig := TestConfig{Name: "test", Value: 42}
	data, _ := json.Marshal(testConfig)
	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Test loading existing file
	var loadedConfig TestConfig
	err = LoadJSONConfig(configPath, &loadedConfig)
	if err != nil {
		t.Errorf("LoadJSONConfig failed: %v", err)
	}

	if loadedConfig.Name != testConfig.Name || loadedConfig.Value != testConfig.Value {
		t.Errorf("Config mismatch. Expected %+v, got %+v", testConfig, loadedConfig)
	}

	// Test loading invalid JSON
	if err := os.WriteFile(configPath, []byte("invalid json"), 0o644); err != nil {
		t.Fatalf("Failed to write invalid JSON: %v", err)
	}

	err = LoadJSONConfig(configPath, &config)
	if err == nil {
		t.Error("LoadJSONConfig should return error for invalid JSON")
	}
}

func TestSaveJSONConfig(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "subdir", "test_config.json")

	type TestConfig struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	testConfig := TestConfig{Name: "test", Value: 42}

	// Test saving config (should create directory if needed)
	err := SaveJSONConfig(configPath, testConfig)
	if err != nil {
		t.Errorf("SaveJSONConfig failed: %v", err)
	}

	// Verify file was created with correct content
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Errorf("Failed to read saved config: %v", err)
	}

	var loadedConfig TestConfig
	if err := json.Unmarshal(data, &loadedConfig); err != nil {
		t.Errorf("Failed to unmarshal saved config: %v", err)
	}

	if loadedConfig.Name != testConfig.Name || loadedConfig.Value != testConfig.Value {
		t.Errorf("Saved config mismatch. Expected %+v, got %+v", testConfig, loadedConfig)
	}

	// Verify JSON is properly formatted (indented)
	if string(data)[0] != '{' || !strings.Contains(string(data), "\n  \"") {
		t.Error("JSON should be properly indented")
	}
}

func TestCheckExecutableExists(t *testing.T) {
	// Test with non-existent file
	if CheckExecutableExists("/non/existent/file") {
		t.Error("CheckExecutableExists should return false for non-existent file")
	}

	// Test with existing file (use the test binary itself)
	execPath, err := os.Executable()
	if err == nil && CheckExecutableExists(execPath) != true {
		t.Error("CheckExecutableExists should return true for existing executable")
	}
}

func TestGetExecutableServerName(t *testing.T) {
	// Test with provided name
	name, err := GetExecutableServerName("custom-name")
	if err != nil || name != "custom-name" {
		t.Errorf("GetExecutableServerName with custom name failed. Expected 'custom-name', got '%s', err: %v", name, err)
	}

	// Test with empty name (should derive from executable)
	name, err = GetExecutableServerName("")
	if err != nil {
		t.Errorf("GetExecutableServerName with empty name failed: %v", err)
	}
	if name == "" {
		t.Error("GetExecutableServerName should return non-empty derived name")
	}
}

func TestGetMCPCommandPath(t *testing.T) {
	tests := []struct {
		name     string
		commands []string
		expected []string
	}{
		{
			name:     "MCPAtRoot",
			commands: []string{"root", MCPCommandName, "post"},
			expected: []string{MCPCommandName},
		},
		{
			name:     "NestedMCP",
			commands: []string{"root", "pre", MCPCommandName, "post"},
			expected: []string{"pre", MCPCommandName},
		},
		{
			name:     "MultipleNestedMCP",
			commands: []string{"root", "pre1", "pre2", MCPCommandName, "post", "post2"},
			expected: []string{"pre1", "pre2", MCPCommandName},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := buildCommand(tt.commands...)
			assert.Equal(t, tt.expected, GetMCPCommandPath(cmd))
		})
	}
}

func buildCommand(cmds ...string) *cobra.Command {
	var parent *cobra.Command
	for _, cmd := range cmds {
		cur := &cobra.Command{Use: cmd}
		if parent != nil {
			parent.AddCommand(cur)
		}
		parent = cur
	}
	return parent
}

func TestDeriveServerName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/path/to/myapp", "myapp"},
		{"/path/to/myapp.exe", "myapp"},
		{"myapp", "myapp"},
		{"my-app.test", "my-app"},
		{"/usr/local/bin/kubectl", "kubectl"},
	}

	for _, test := range tests {
		result := DeriveServerName(test.input)
		if result != test.expected {
			t.Errorf("DeriveServerName(%s) = %s, expected %s", test.input, result, test.expected)
		}
	}
}
