// Package bridge provides functionality to convert Cobra CLI applications into MCP servers.
// It handles the registration of Cobra commands as MCP tools and manages command execution.
package bridge

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
)

// CommandExecFunc is a function type that executes a command and returns an MCP tool result.
type CommandExecFunc func(ctx context.Context) *mcp.CallToolResult

// CommandFactory is an interface for creating Cobra commands for registration and execution.
// It provides a factory pattern to ensure fresh command instances for each execution,
// preventing state pollution between different MCP tool calls.
//
// The factory should implement two methods:
// - RegistrationCommand(): Returns a command tree used for MCP tool registration only
// - New(): Returns a fresh command instance and execution function for each tool call
type CommandFactory interface {
	// RegistrationCommand returns a Cobra command tree for MCP tool registration.
	// This command is used to discover the available commands and their flags.
	// It should not be executed directly.
	RegistrationCommand() *cobra.Command

	// New creates a fresh command instance and returns both the command and
	// an execution function. This ensures clean state for each tool call.
	New() (*cobra.Command, CommandExecFunc)
}

// CobraToMCPBridge converts a Cobra CLI application to an MCP server.
// The bridge is thread-safe for concurrent MCP tool calls as it creates
// fresh command instances for each execution via the CommandFactory.
type CobraToMCPBridge struct {
	commandFactory CommandFactory    // Factory function to create fresh command instances
	server         *server.MCPServer // The MCP server instance
	logger         *slog.Logger
}

// NewCobraToMCPBridge creates a new bridge instance with validation
func NewCobraToMCPBridge(factory CommandFactory, config *MCPCommandConfig) (*CobraToMCPBridge, error) {
	if factory == nil {
		return nil, fmt.Errorf("cmdFactory cannot be nil")
	}
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	if config.AppName == "" {
		return nil, fmt.Errorf("appName cannot be empty")
	}
	if config.AppVersion == "" {
		config.AppVersion = "unknown"
	}

	logger, err := config.NewSlogger()
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	b := &CobraToMCPBridge{
		commandFactory: factory,
		logger:         logger,
		server: server.NewMCPServer(
			config.AppName,
			config.AppVersion,
		),
	}

	registrationCmd, err := func() (cmd *cobra.Command, err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("registration command panicked: %v", r)
				cmd = nil
			}
		}()
		cmd = b.commandFactory.RegistrationCommand()
		return cmd, err
	}()
	if err != nil {
		return nil, fmt.Errorf("failed to get registration command: %w", err)
	}
	if registrationCmd == nil {
		return nil, fmt.Errorf("registration command cannot be nil")
	}

	b.registerCommands(registrationCmd, "")
	return b, nil
}

// StartServer starts the MCP server using stdio transport
func (b *CobraToMCPBridge) StartServer() error {
	return server.ServeStdio(b.server)
}
