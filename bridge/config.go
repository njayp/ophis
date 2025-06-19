package bridge

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// Config holds configuration for the MCP command
type Config struct {
	AppName    string
	AppVersion string
	LogFile    string
	LogLevel   string
}

// newSlogger makes a new slog.logger that writes to file. Don't give the user
// the option to write to stdout, because that causes errors.
func (c *Config) newSlogger() (*slog.Logger, error) {
	// if logfile not set, use usercache
	if c.LogFile == "" {
		// Get the cache directory
		cacheDir, err := os.UserCacheDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user cache directory: %w", err)
		}

		// Create the log directory
		logDir := filepath.Join(cacheDir, "mcp-servers", c.AppName, "logs")
		if err := os.MkdirAll(logDir, 0o700); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}

		c.LogFile = filepath.Join(logDir, "server.log")
	}

	file, err := os.OpenFile(c.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	// Create handler options
	opts := &slog.HandlerOptions{
		Level: parseLogLevel(c.LogLevel),
		//AddSource: true,
	}

	// Create handler based on format preference
	handler := slog.NewTextHandler(file, opts)
	return slog.New(handler), nil
}

// parseLogLevel converts a string log level to slog.Level.
// Supported levels are: debug, info, warn, error (case-insensitive).
// Defaults to info for unknown levels.
func parseLogLevel(level string) slog.Level {
	// Parse log level
	slogLevel := slog.LevelInfo
	switch strings.ToLower(level) {
	case "debug":
		slogLevel = slog.LevelDebug
	case "info":
		slogLevel = slog.LevelInfo
	case "warn":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	}

	return slogLevel
}
