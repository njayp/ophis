// Package bridge provides functionality to convert Cobra CLI applications into MCP servers.
// It handles the registration of Cobra commands as MCP tools and manages command execution.
package bridge

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/njayp/ophis/tools"
	"github.com/spf13/cobra"
)

// CommandExecFunc is a function type that executes a command and returns an MCP tool result.
type CommandExecFunc func(context.Context, *cobra.Command) *mcp.CallToolResult

// CommandFactory is an interface for creating Cobra commands for registration and execution.
// It provides a factory pattern to ensure fresh command instances for each execution,
// preventing state pollution between different MCP tool calls.
//
// Implementation Requirements:
// - Tools(): Must return a stable list of tools derived from your command tree
// - New(): Must create completely fresh command instances on each call
type CommandFactory interface {
	// Tools returns all available MCP tools from your command tree.
	// This should return a consistent list based on your application's command structure.
	Tools() []tools.Tool

	// New creates a fresh command instance and returns both the command and
	// an execution function. This ensures clean state for each tool call.
	// The returned command should be a completely new instance to prevent
	// state pollution between concurrent executions.
	New() (*cobra.Command, CommandExecFunc)
}

// Manager converts a Cobra CLI application to an MCP server.
// The bridge is thread-safe for concurrent MCP tool calls as it creates
// fresh command instances for each execution via the CommandFactory.
type Manager struct {
	commandFactory CommandFactory    // Factory function to create fresh command instances
	server         *server.MCPServer // The MCP server instance
	logger         *slog.Logger
}

// New creates a new bridge instance with validation
func New(factory CommandFactory, config *Config) (*Manager, error) {
	if factory == nil {
		return nil, fmt.Errorf("command factory cannot be nil: must provide a CommandFactory implementation")
	}
	if config == nil {
		return nil, fmt.Errorf("configuration cannot be nil: must provide a Config struct with AppName and AppVersion")
	}
	if config.AppName == "" {
		return nil, fmt.Errorf("application name cannot be empty: Config.AppName is required for server identification")
	}
	if config.AppVersion == "" {
		config.AppVersion = "unknown"
	}

	logger, err := config.newSlogger()
	if err != nil {
		return nil, fmt.Errorf("failed to create logger for MCP server: %w", err)
	}

	b := &Manager{
		commandFactory: factory,
		logger:         logger,
		server: server.NewMCPServer(
			config.AppName,
			config.AppVersion,
			config.ServerOptions...,
		),
	}

	tools, err := func() (tools []tools.Tool, err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("command factory Tools() method panicked during tool registration: %v", r)
				tools = nil
			}
		}()
		tools = b.commandFactory.Tools()
		return tools, err
	}()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve MCP tools from command factory: %w", err)
	}

	b.registerTools(tools)
	return b, nil
}

// StartServer starts the MCP server using stdio transport
func (b *Manager) StartServer() error {
	return server.ServeStdio(b.server)
}

// GetLogger returns the logger instance used by the bridge
func (b *Manager) GetLogger() *slog.Logger {
	return b.logger
}
