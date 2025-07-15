package bridge

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/mark3labs/mcp-go/server"
)

// Config holds configuration for the MCP command
type Config struct {
	AppName       string
	AppVersion    string
	LogFile       string
	LogLevel      string
	ServerOptions []server.ServerOption
}

// newSlogger makes a new slog.logger that writes to file. Don't give the user
// the option to write to stdout, because that causes errors.
func (c *Config) newSlogger() (*slog.Logger, error) {
	path, err := c.logFilePath()
	if err != nil {
		return nil, fmt.Errorf("failed to determine log file path: %w", err)
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file at '%s': %w", c.LogFile, err)
	}

	// Create handler options
	opts := &slog.HandlerOptions{
		Level: parseLogLevel(c.LogLevel),
		// AddSource: true,
	}

	// Create handler based on format preference
	handler := slog.NewTextHandler(file, opts)
	logger := slog.New(handler)

	// Signal handling
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Close logger gracefully
	go func() {
		<-quit
		logger.Info("Closing logger")
		if err := file.Close(); err != nil {
			logger.Error("Error closing log file:", "error", err)
		}
	}()

	return logger, nil
}

func (c *Config) logFilePath() (string, error) {
	if c.LogFile != "" {
		return c.LogFile, nil
	}

	// Get the cache directory
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user cache directory for log file storage: %w", err)
	}

	// Create the log directory
	logDir := filepath.Join(cacheDir, "mcp-servers", c.AppName, "logs")
	if err := os.MkdirAll(logDir, 0o700); err != nil {
		return "", fmt.Errorf("failed to create log directory at '%s': %w", logDir, err)
	}

	return filepath.Join(logDir, "server.log"), nil
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
