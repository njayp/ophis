package bridge

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// Constants for MCP parameter names and error messages
const (
	MCPCommandName = "mcp"
	// PositionalArgsParam is the parameter name for positional arguments
	PositionalArgsParam = "args"
	FlagsParam          = "flags"
)

// MCPCommandConfig holds configuration for the MCP command
type MCPCommandConfig struct {
	AppName    string
	AppVersion string
	LogFile    string
	LogLevel   string
}

// Validate checks if the configuration is valid
func (c *MCPCommandConfig) Validate() error {
	if c == nil {
		return fmt.Errorf("config cannot be nil")
	}
	if c.AppName == "" {
		return fmt.Errorf("app name cannot be empty")
	}
	// LogLevel and LogFile are optional, so no validation needed
	return nil
}

// NewSlogger makes a new slog.logger that writes to file. Don't give the user
// the option to write to stdout, because that causes errors.
func (c *MCPCommandConfig) NewSlogger() (*slog.Logger, error) {
	if err := c.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
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
		Level:     ParseLogLevel(c.LogLevel),
		AddSource: true,
	}

	// Create handler based on format preference
	handler := slog.NewTextHandler(file, opts)
	return slog.New(handler), nil
}

// ParseLogLevel converts a string log level to slog.Level.
// Supported levels are: debug, info, warn, error (case-insensitive).
// Defaults to info for unknown levels.
func ParseLogLevel(level string) slog.Level {
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
