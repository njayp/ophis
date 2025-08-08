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

// Manager manages the bridge between a Cobra CLI application and an MCP server.
// It handles tool registration, command execution, and server lifecycle.
//
// Manager instances should be created using the New function rather than
// direct struct initialization to ensure proper validation and setup.
type Manager struct {
	server *server.MCPServer // The underlying MCP server instance
}

// New creates a new Manager instance from the provided configuration.
// Returns an error if:
//   - config is nil
//   - config.RootCmd is nil
func New(config *Config) (*Manager, error) {
	if config == nil {
		return nil, fmt.Errorf("configuration cannot be nil: must provide a Config struct with a RootCmd")
	}

	if config.RootCmd == nil {
		return nil, fmt.Errorf("root command cannot be nil: Config.RootCmd is required to register tools")
	}

	config.setupSlogger()

	appName := config.RootCmd.Name()
	slog.Info("creating MCP server", "app_name", appName, "app_version", config.RootCmd.Version)

	opts := append(config.ServerOptions, server.WithRecovery())
	server := server.NewMCPServer(
		appName,
		config.RootCmd.Version,
		opts...,
	)

	b := &Manager{
		server: server,
	}

	b.registerTools(config.Tools())
	return b, nil
}

// StartServer starts the MCP server using stdio transport.
//
// This method blocks until the server is shut down or encounters an error.
// The server communicates over stdin/stdout, making it compatible with
// MCP clients like Claude Desktop.
func (b *Manager) StartServer() error {
	return server.ServeStdio(b.server)
}
