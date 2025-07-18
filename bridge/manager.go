// Package bridge provides functionality to convert Cobra CLI applications into MCP servers.
// It handles the registration of Cobra commands as MCP tools and manages command execution.
package bridge

import (
	"fmt"
	"log/slog"

	"github.com/mark3labs/mcp-go/server"
	"github.com/njayp/ophis/tools"
)

// ToolsGenerator is a function type that generates MCP tools from Cobra commands.
type ToolsGenerator func() []tools.Tool

// Manager converts a Cobra CLI application to an MCP server.
// The bridge is thread-safe for concurrent MCP tool calls as it creates
// fresh command instances for each execution via the CommandFactory.
type Manager struct {
	server *server.MCPServer // The MCP server instance
	logger *slog.Logger
}

// New creates a new bridge instance with validation
func New(config *Config) (*Manager, error) {
	if config == nil {
		return nil, fmt.Errorf("configuration cannot be nil: must provide a Config struct with AppName and AppVersion")
	}

	if config.AppName == "" {
		return nil, fmt.Errorf("application name cannot be empty: Config.AppName is required for server identification")
	}

	if config.RootCmd == nil {
		return nil, fmt.Errorf("root command cannot be nil: Config.RootCmd is required to register tools")
	}

	if config.AppVersion == "" {
		config.AppVersion = "unknown"
	}

	logger := config.newSlogger()
	logger.Info("Creating MCP server", "app_name", config.AppName, "app_version", config.AppVersion)

	opts := append(config.ServerOptions, server.WithRecovery())
	server := server.NewMCPServer(
		config.AppName,
		config.AppVersion,
		opts...,
	)

	b := &Manager{
		logger: logger,
		server: server,
	}

	b.registerTools(config.Tools())
	return b, nil
}

// StartServer starts the MCP server using stdio transport
func (b *Manager) StartServer() error {
	return server.ServeStdio(b.server)
}
