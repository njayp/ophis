package ophis

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Constants for MCP parameter names and error messages
const (
	// PositionalArgsParam is the parameter name for positional arguments
	PositionalArgsParam = "args"

	// Error messages
	ErrNoRunFunction    = "Command has no Run or RunE function"
	ErrArrayUnsupported = "Array types are not supported in MCP tools"
)

// CobraToMCPBridge converts a Cobra CLI application to an MCP server
type CobraToMCPBridge struct {
	commandFactory func() *cobra.Command // Factory function to create fresh command instances
	appName        string
	version        string
	server         *server.MCPServer // The MCP server instance
	logger         *slog.Logger
}

// NewCobraToMCPBridge creates a new bridge instance with validation
func NewCobraToMCPBridge(cmdFactory func() *cobra.Command, appName, version string, logger *slog.Logger) *CobraToMCPBridge {
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

	return &CobraToMCPBridge{
		commandFactory: cmdFactory,
		appName:        appName,
		version:        version,
		server:         server.NewMCPServer(appName, version),
		logger:         logger,
	}
}

// CreateMCPServer creates and configures the MCP server with tools for each Cobra command
func (b *CobraToMCPBridge) CreateMCPServer() *server.MCPServer {
	b.logger.Info("Creating MCP server", "app_name", b.appName, "version", b.version)

	s := server.NewMCPServer(
		b.appName,
		b.version,
	)

	b.registerCommands(b.commandFactory(), "")
	return s
}

// registerCommands recursively registers all Cobra commands as MCP tools
func (b *CobraToMCPBridge) registerCommands(cmd *cobra.Command, parentPath string) {
	// Create the tool name
	toolName := cmd.Name()
	if parentPath != "" {
		toolName = parentPath + "_" + cmd.Name()
	}

	// Register subcommands
	for _, subCmd := range cmd.Commands() {
		if subCmd.Hidden {
			continue
		}
		subPath := toolName
		if parentPath == "" && cmd.Name() != toolName {
			subPath = cmd.Name()
		}
		b.logger.Debug("Registering subcommand", "name", subCmd.Name(), "path", subPath)
		b.registerCommands(subCmd, subPath)
	}

	// Skip if the command has no runnable function
	if cmd.Run == nil && cmd.RunE == nil {
		return
	}

	// Create MCP tool options
	toolOptions := []mcp.ToolOption{
		mcp.WithDescription(b.getCommandDescription(cmd, parentPath)),
	}

	// Add parameters for flags (both local and persistent)
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		if flag.Hidden {
			return
		}
		toolOptions = append(toolOptions, b.createParameterFromFlag(flag)...)
	})

	// Add parameters for persistent flags from parent commands
	cmd.PersistentFlags().VisitAll(func(flag *pflag.Flag) {
		if flag.Hidden {
			return
		}
		// Check if this flag was already added from local flags to avoid duplicates
		if cmd.Flags().Lookup(flag.Name) == nil {
			toolOptions = append(toolOptions, b.createParameterFromFlag(flag)...)
		}
	})

	// Add inherited persistent flags from parent commands
	for parent := cmd.Parent(); parent != nil; parent = parent.Parent() {
		parent.PersistentFlags().VisitAll(func(flag *pflag.Flag) {
			if flag.Hidden {
				return
			}
			// Check if this flag was already added to avoid duplicates
			if cmd.Flags().Lookup(flag.Name) == nil && cmd.PersistentFlags().Lookup(flag.Name) == nil {
				toolOptions = append(toolOptions, b.createParameterFromFlag(flag)...)
			}
		})
	}

	// Add an "args" parameter for positional arguments
	// This allows MCP clients to pass positional arguments that aren't flags
	argsDescription := "Space-separated positional arguments for the command"
	if cmd.Use != "" {
		argsDescription += fmt.Sprintf(". Usage: %s", cmd.Use)
	}
	if cmd.Args != nil {
		// Try to provide more specific information about expected arguments
		argsDescription += ". See command usage for argument requirements."
	}
	toolOptions = append(toolOptions, mcp.WithString(PositionalArgsParam,
		mcp.Description(argsDescription),
	))

	// Create the tool
	tool := mcp.NewTool(toolName, toolOptions...)
	b.logger.Debug("Registering MCP tool", "tool_name", toolName, "command", cmd.Name())

	// Add the tool handler
	b.server.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		b.logger.Info("MCP tool request received", "tool_name", toolName, "arguments", request.Params.Arguments)

		// TODO execute new command instance to avoid state pollution
		freshCmd, err := b.findCommandInTree(b.commandFactory(), cmd, parentPath)
		if err != nil {
			b.logger.Error("Failed to create subcommand", "tool_name", toolName, "error", err)
			return nil, fmt.Errorf("failed to create subcommand: %w", err)
		}

		result, err := b.executeCommand(ctx, freshCmd, parentPath, request)
		if err != nil {
			b.logger.Error("MCP tool execution failed", "tool_name", toolName, "error", err)
		} else {
			b.logger.Info("MCP tool executed successfully", "tool_name", toolName)
		}
		return result, err
	})

}

// getCommandDescription creates a description for the MCP tool from the Cobra command
func (b *CobraToMCPBridge) getCommandDescription(cmd *cobra.Command, parentPath string) string {
	desc := cmd.Short
	if desc == "" {
		desc = cmd.Long
	}
	if desc == "" {
		cmdPath := cmd.Name()
		if parentPath != "" {
			cmdPath = strings.ReplaceAll(parentPath, "_", " ") + " " + cmd.Name()
		}
		desc = fmt.Sprintf("Execute the '%s' command", cmdPath)
	}

	// Add usage information
	if cmd.Use != "" {
		desc += fmt.Sprintf("\n\nUsage: %s", cmd.Use)
	}

	if cmd.Long != "" && cmd.Long != cmd.Short {
		desc += fmt.Sprintf("\n\n%s", cmd.Long)
	}

	b.logger.Debug("Command description", "name", cmd.Name(), "description", desc)
	return desc
}

// createParameterFromFlag converts a Cobra flag to MCP tool parameters
func (b *CobraToMCPBridge) createParameterFromFlag(flag *pflag.Flag) []mcp.ToolOption {
	flagName := strings.ReplaceAll(flag.Name, "-", "_")
	description := flag.Usage
	if description == "" {
		description = fmt.Sprintf("Flag: %s", flag.Name)
	}

	// Use helper function to determine the correct MCP parameter type
	return b.createMCPParameter(flagName, description, flag.Value.Type())
}

// createMCPParameter creates an MCP parameter based on the type
func (b *CobraToMCPBridge) createMCPParameter(name, description, flagType string) []mcp.ToolOption {
	switch flagType {
	case "bool":
		return []mcp.ToolOption{mcp.WithBoolean(name, mcp.Description(description))}
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "float32", "float64":
		return []mcp.ToolOption{mcp.WithNumber(name, mcp.Description(description))}
	case "string":
		return []mcp.ToolOption{mcp.WithString(name, mcp.Description(description))}
	default:
		// Default to string for unknown types
		return []mcp.ToolOption{mcp.WithString(name, mcp.Description(description))}
	}
}

// findCommandInTree finds the equivalent command in a fresh command tree
func (b *CobraToMCPBridge) findCommandInTree(root *cobra.Command, target *cobra.Command, parentPath string) (*cobra.Command, error) {
	if parentPath == "" {
		return root, nil // If no parent path, return the root command
	}

	// Navigate to the parent path first
	pathParts := strings.Split(parentPath, "_")
	current := root

	// Skip the first part if it matches the root command name
	if len(pathParts) > 0 && pathParts[0] == root.Name() {
		pathParts = pathParts[1:]
	}

	for _, part := range pathParts {
		found := false
		for _, cmd := range current.Commands() {
			if cmd.Name() == part {
				current = cmd
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("parent path part %s not found in fresh tree", part)
		}
	}

	// Now find the target command
	for _, cmd := range current.Commands() {
		if cmd.Name() == target.Name() {
			return cmd, nil
		}
	}

	return nil, fmt.Errorf("command %s not found under parent %s in fresh tree", target.Name(), parentPath)
}

// executeCommand executes the Cobra command using a fresh instance to avoid state pollution
func (b *CobraToMCPBridge) executeCommand(ctx context.Context, cmd *cobra.Command, parentPath string, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.GetArguments()
	var args []string

	// Handle positional arguments from the "args" parameter
	if argsValue, exists := arguments[PositionalArgsParam]; exists {
		if argsStr, ok := argsValue.(string); ok && argsStr != "" {
			// Split the args string by spaces to get individual arguments
			args = strings.Fields(argsStr)
			b.logger.Debug("Parsed positional arguments", "args", args)
		}
	}

	// Helper function to process flag values with improved error handling
	processFlag := func(flag *pflag.Flag) {
		if flag.Hidden {
			return
		}

		// Skip the special "args" parameter as it's not a flag
		if flag.Name == PositionalArgsParam {
			return
		}
		value, exists := arguments[flag.Name]
		if !exists {
			return
		}

		b.logger.Debug("Setting flag", "flag_name", flag.Name, "value", value)

		err := flag.Value.Set(fmt.Sprintf("%v", value)) // Use fmt.Sprintf to handle different types
		if err != nil {
			b.logger.Error("Invalid flag value", "flag_name", flag.Name, "error", err)
		}
	}

	// Process flags from MCP arguments and set them directly on the original command
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		processFlag(flag)
	})

	// Process persistent flags
	cmd.PersistentFlags().VisitAll(func(flag *pflag.Flag) {
		processFlag(flag)
	})

	// Create isolated output capture to avoid interfering with MCP protocol
	var output strings.Builder
	cmd.SetOut(&output)
	cmd.SetErr(&output)

	// Execute the command's Run function with proper error recovery
	var err error
	func() {
		// Recover from any panics that might occur during command execution
		defer func() {
			if r := recover(); r != nil {
				b.logger.Error("Command execution panicked", "command", cmd.Name(), "panic", r)
				err = fmt.Errorf("command panicked: %v", r)
			}
		}()

		err = cmd.Execute()
	}()

	if err != nil {
		b.logger.Error("Failed to execute command function", "command", cmd.Name(), "error", err, "output", output.String())
		return mcp.NewToolResultError(fmt.Sprintf("Command failed: %v\nOutput: %s", err, output.String())), nil
	}

	result := output.String()
	b.logger.Debug("Command executed successfully", "command", cmd.Name(), "output", result)
	return mcp.NewToolResultText(result), nil
}

// StartServer starts the MCP server using stdio transport
func (b *CobraToMCPBridge) StartServer() error {
	return server.ServeStdio(b.server)
}
