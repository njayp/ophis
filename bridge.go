package ophis

import (
	"context"
	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
)

// Constants for MCP parameter names and error messages
const (
	// PositionalArgsParam is the parameter name for positional arguments
	PositionalArgsParam = "args"
	FlagsParam          = "flags"
)

type CommandExecFunc func(ctx context.Context) *mcp.CallToolResult

type CommandFactory interface {
	CreateRegistrationCommand() *cobra.Command
	CreateCommand() (*cobra.Command, CommandExecFunc)
}

// MCPCommandConfig holds configuration for the MCP command
type MCPCommandConfig struct {
	AppName    string
	AppVersion string
	Logger     *slog.Logger
}

// CobraToMCPBridge converts a Cobra CLI application to an MCP server
type CobraToMCPBridge struct {
	commandFactory CommandFactory    // Factory function to create fresh command instances
	server         *server.MCPServer // The MCP server instance
	logger         *slog.Logger
}

// NewCobraToMCPBridge creates a new bridge instance with validation
func NewCobraToMCPBridge(cmdFactory CommandFactory, config *MCPCommandConfig) *CobraToMCPBridge {
	if cmdFactory == nil {
		panic("cmdFactory cannot be nil")
	}
	if config.AppName == "" {
		panic("appName cannot be empty")
	}
	if config.AppVersion == "" {
		config.AppVersion = "unknown"
	}
	if config.Logger == nil {
		config.Logger = slog.Default()
	}

	b := &CobraToMCPBridge{
		commandFactory: cmdFactory,
		logger:         config.Logger,
		server: server.NewMCPServer(
			config.AppName,
			config.AppVersion,
		),
	}

	b.registerCommands(b.commandFactory.CreateRegistrationCommand(), "")
	return b
}

// StartServer starts the MCP server using stdio transport
func (b *CobraToMCPBridge) StartServer() error {
	return server.ServeStdio(b.server)
}
