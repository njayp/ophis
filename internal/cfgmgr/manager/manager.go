package manager

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Platform[T, S any] interface {
	ConfigPath() string
	AddServer(config *T, name string, server S)
	RemoveServer(config *T, name string)
}

type Manager[T, S any] struct {
	Platform Platform[T, S]
}

func (m *Manager[T, S]) EnableMCPServer(name string, server S) error {
	var config T
	err := m.LoadJSONConfig(&config)
	if err != nil {
		return err
	}

	m.Platform.AddServer(&config, name, server)
	return m.SaveJSONConfig(&config)
}

func (m *Manager[T, S]) DisableMCPServer(name string) error {
	var config T
	err := m.LoadJSONConfig(&config)
	if err != nil {
		return err
	}

	m.Platform.RemoveServer(&config, name)
	return m.SaveJSONConfig(&config)
}

// LoadJSONConfig unmarshals a JSON file into the provided interface.
func (m *Manager[T, S]) LoadJSONConfig(config *T) error {
	path := m.Platform.ConfigPath()

	// Check if config file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// File doesn't exist - caller should handle initialization
		return nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read configuration file at %q: %w", path, err)
	}

	if err := json.Unmarshal(data, config); err != nil {
		return fmt.Errorf("failed to parse configuration file at %q: invalid JSON format: %w", path, err)
	}

	return nil
}

// SaveJSONConfig marshals and saves configuration as formatted JSON.
func (m *Manager[T, S]) SaveJSONConfig(config *T) error {
	path := m.Platform.ConfigPath()

	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("failed to create configuration directory at %q: %w", filepath.Dir(path), err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal configuration to JSON: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("failed to write configuration file at %q: %w", path, err)
	}

	return nil
}
