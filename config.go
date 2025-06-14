package ophis

import (
	"io"
	"log/slog"
	"strings"
)

// Constants for MCP parameter names and error messages
const (
	// PositionalArgsParam is the parameter name for positional arguments
	PositionalArgsParam = "args"
	FlagsParam          = "flags"
)

// MCPCommandConfig holds configuration for the MCP command
type MCPCommandConfig struct {
	AppName    string
	AppVersion string
	LogOut     io.Writer
	LogLevel   string
}

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
