package ophis

import (
	"log/slog"
	"strings"
)

// parseLogLevel converts a string log level to slog.Level.
// Supported levels: debug, info, warn, error (case-insensitive).
// Returns slog.LevelInfo for unrecognized values.
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
