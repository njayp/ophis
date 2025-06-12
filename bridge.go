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

	b := &CobraToMCPBridge{
		commandFactory: cmdFactory,
		appName:        appName,
		version:        version,
		logger:         logger,
		server: server.NewMCPServer(
			appName,
			version,
		),
	}

	b.registerCommands(b.commandFactory(), "")
	return b
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
		b.registerCommands(subCmd, subPath)
	}

	// Skip if the command has no runnable function
	if cmd.Run == nil && cmd.RunE == nil {
		return
	}

	b.logger.Debug("Registering command", "name", cmd.Name(), "path", parentPath)
	// Create MCP tool options
	toolOptions := []mcp.ToolOption{
		mcp.WithDescription(b.getCommandDescription(cmd, parentPath)),
	}

	// Add parameters for flags (both local and persistent)
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		if flag.Hidden {
			return
		}
		b.logger.Debug("Registering Tool Parameter", slog.String("cmd", cmd.Name()), slog.String("name", flag.Name))
		toolOptions = append(toolOptions, b.createParameterFromFlag(flag)...)
	})

	// Add all persistent flags recursively
	for t := cmd; t != nil; t = t.Parent() {
		// Add parameters for persistent flags from parent commands
		t.PersistentFlags().VisitAll(func(flag *pflag.Flag) {
			if flag.Hidden {
				return
			}
			// Check if this flag was already added from local flags to avoid duplicates
			if cmd.Flags().Lookup(flag.Name) == nil {
				b.logger.Debug("Registering Tool Parameter", slog.String("cmd", cmd.Name()), slog.String("name", flag.Name))
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
		result := b.executeCommand(ctx, cmd, request)
		// TODO figure out what err is used for
		return result, nil
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
	description := flag.Usage
	if description == "" {
		description = fmt.Sprintf("Flag: %s", flag.Name)
	}

	// Use helper function to determine the correct MCP parameter type
	return b.createMCPParameter(flag.Name, description, flag.Value.Type())
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

// executeCommand executes the Cobra command using a fresh instance to avoid state pollution
func (b *CobraToMCPBridge) executeCommand(ctx context.Context, cmd *cobra.Command, request mcp.CallToolRequest) *mcp.CallToolResult {
	arguments := request.GetArguments()
	var args []string

	// Add command path args
	if cmd.Parent() != nil {
		args = append(args, strings.Fields(cmd.CommandPath())[1:]...)
	}

	// Handle positional arguments from the "args" parameter
	if argsValue, exists := arguments[PositionalArgsParam]; exists {
		if argsStr, ok := argsValue.(string); ok && argsStr != "" {
			// Split the args string by spaces to get individual arguments
			args = append(args, strings.Fields(argsStr)...)
			b.logger.Debug("Parsed positional arguments", "args", args)
		}
	}

	// get a fresh command struct for each execution
	cmd = b.commandFactory()
	cmd.SetArgs(args)
	b.logger.Debug("Set command arguments", "args", args)

	for k, v := range arguments {
		if k == PositionalArgsParam {
			continue
		}

		b.logger.Debug("setting flag", slog.String("cmd", cmd.Name()), slog.String("flag", k))
		flag := cmd.Flag(k)
		if flag == nil {
			// TODO maybe return err
			continue
		}

		// TODO verify that v is the right type
		str, ok := v.(string)
		if !ok {
			b.logger.Error("Failed to load flag", slog.String("cmd", cmd.Name()), slog.String("key", k), slog.Any("value", v))
			return mcp.NewToolResultError(fmt.Sprintf("%s: failed to load flag %s", cmd.Name(), k))
		}
		err := flag.Value.Set(str)
		if err != nil {
			b.logger.Error("Failed to load flag", slog.String("cmd", cmd.Name()), slog.String("key", k), slog.Any("value", v))
			return mcp.NewToolResultErrorFromErr(fmt.Sprintf("%s: failed to load flag %s", cmd.Name(), k), err)
		}
	}

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

		err = cmd.ExecuteContext(ctx)
	}()

	if err != nil {
		b.logger.Error("Failed to execute command function", "command", cmd.Name(), "error", err, "output", output.String())
		return mcp.NewToolResultError(fmt.Sprintf("Command failed: %v\nOutput: %s", err, output.String()))
	}

	result := output.String()
	b.logger.Debug("Command executed successfully", "command", cmd.Name(), "output", result)
	return mcp.NewToolResultText(result)
}

// StartServer starts the MCP server using stdio transport
func (b *CobraToMCPBridge) StartServer() error {
	return server.ServeStdio(b.server)
}
