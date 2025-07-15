package bridge

import (
	"log/slog"
	"testing"
)

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"DEBUG", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"INFO", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"WARN", slog.LevelWarn},
		{"error", slog.LevelError},
		{"ERROR", slog.LevelError},
		{"unknown", slog.LevelInfo}, // Default
		{"", slog.LevelInfo},        // Default
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseLogLevel(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestLogFilePath(t *testing.T) {
	config := &Config{
		AppName: "testapp",
	}

	// Test with custom log file path
	config.LogFile = "/tmp/testapp.log"
	path, err := config.logFilePath()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if path != config.LogFile {
		t.Errorf("Expected log file path %s, got %s", config.LogFile, path)
	}

	// Test without custom log file path
	config.LogFile = ""
	path, err = config.logFilePath()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if path == "" {
		t.Error("Expected a valid log file path, got empty string")
	}
}
