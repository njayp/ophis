package cfgmgr

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestBackupConfigFile(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.json")

	t.Run("non-existent file", func(t *testing.T) {
		err := BackupConfigFile(configPath)
		assert.NoError(t, err, "BackupConfigFile should return nil for non-existent file")
	})

	t.Run("create first backup", func(t *testing.T) {
		testData := []byte(`{"test": "data"}`)
		err := os.WriteFile(configPath, testData, 0o644)
		assert.NoError(t, err)

		err = BackupConfigFile(configPath)
		assert.NoError(t, err)

		backupPath := configPath + ".backup"
		backupData, err := os.ReadFile(backupPath)
		assert.NoError(t, err)
		assert.Equal(t, string(testData), string(backupData))
	})

	t.Run("rotate backups", func(t *testing.T) {
		// Create multiple backups to test rotation
		for i := 1; i <= MaxBackups+2; i++ {
			testData := []byte(fmt.Sprintf(`{"version": %d}`, i))
			err := os.WriteFile(configPath, testData, 0o644)
			assert.NoError(t, err)

			err = BackupConfigFile(configPath)
			assert.NoError(t, err)
		}

		// Check that we have MaxBackups files (not counting the main config)
		// Should have: .backup, .backup.1, .backup.2, .backup.3, .backup.4
		for i := 0; i < MaxBackups; i++ {
			var backupPath string
			if i == 0 {
				backupPath = configPath + ".backup"
			} else {
				backupPath = fmt.Sprintf("%s.backup.%d", configPath, i)
			}
			_, err := os.Stat(backupPath)
			assert.NoError(t, err, "Backup %s should exist", backupPath)
		}

		// Check that older backups were removed
		oldBackup := fmt.Sprintf("%s.backup.%d", configPath, MaxBackups)
		_, err := os.Stat(oldBackup)
		assert.True(t, os.IsNotExist(err), "Old backup %s should not exist", oldBackup)
	})
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
			result, err := GetCmdPath(cmd)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
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
