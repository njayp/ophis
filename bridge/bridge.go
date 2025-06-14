package bridge

import (
	"context"
	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
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
func NewCobraToMCPBridge(factory CommandFactory, config *MCPCommandConfig) *CobraToMCPBridge {
	if factory == nil {
		panic("cmdFactory cannot be nil")
	}
	if config.AppName == "" {
		panic("appName cannot be empty")
	}
	if config.AppVersion == "" {
		config.AppVersion = "unknown"
	}

	logger, err := config.NewSlogger()
	if err != nil {
		panic(err)
	}

	b := &CobraToMCPBridge{
		commandFactory: factory,
		logger:         logger,
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
