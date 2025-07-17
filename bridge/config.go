package bridge

import (
	"log/slog"
	"os"
	"strings"

	"github.com/mark3labs/mcp-go/server"
	"github.com/njayp/ophis/tools"
	"github.com/spf13/cobra"
)

// Config holds configuration for the MCP command
type Config struct {
	AppName       string
	AppVersion    string
	LogFile       string
	LogLevel      string
	RootCmd       *cobra.Command // The root command for the MCP server
	Generator     *tools.Generator
	ServerOptions []server.ServerOption
}

func (c *Config) Tools() []tools.Tool {
	if c.Generator != nil {
		return c.Generator.FromRootCmd(c.RootCmd)
	}
	return tools.FromRootCmd(c.RootCmd)
}

// newSlogger makes a new slog.logger that writes to file. Don't give the user
// the option to write to stdout, because that causes errors.
func (c *Config) newSlogger() *slog.Logger {
	// Create handler options
	opts := &slog.HandlerOptions{
		Level: parseLogLevel(c.LogLevel),
		// AddSource: true,
	}

	// Create handler based on format preference
	handler := slog.NewTextHandler(os.Stderr, opts)
	return slog.New(handler)
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
