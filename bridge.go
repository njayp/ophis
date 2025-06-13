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

// CobraToMCPBridge converts a Cobra CLI application to an MCP server
type CobraToMCPBridge struct {
	commandFactory CommandFactory    // Factory function to create fresh command instances
	server         *server.MCPServer // The MCP server instance
	logger         *slog.Logger
}

// NewCobraToMCPBridge creates a new bridge instance with validation
func NewCobraToMCPBridge(cmdFactory CommandFactory, appName, version string, logger *slog.Logger) *CobraToMCPBridge {
	if cmdFactory == nil {
		panic("cmdFactory cannot be nil")
	}
	if appName == "" {
		panic("appName cannot be empty")
	}
	if version == "" {
		version = "unknown"
	}
	if logger == nil {
		logger = slog.Default()
	}

	b := &CobraToMCPBridge{
		commandFactory: cmdFactory,
		logger:         logger,
		server: server.NewMCPServer(
			appName,
			version,
		),
	}

	b.registerCommands(b.commandFactory.CreateRegistrationCommand(), "")
	return b
}

// StartServer starts the MCP server using stdio transport
func (b *CobraToMCPBridge) StartServer() error {
	return server.ServeStdio(b.server)
}
