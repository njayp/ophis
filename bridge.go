package ophis

import (
	"context"
	"fmt"
	"io"
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

type CommandFactory func(output io.Writer) *cobra.Command

// CobraToMCPBridge converts a Cobra CLI application to an MCP server
type CobraToMCPBridge struct {
	commandFactory CommandFactory // Factory function to create fresh command instances
	appName        string
	version        string
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
		appName:        appName,
		version:        version,
		logger:         logger,
		server: server.NewMCPServer(
			appName,
			version,
		),
	}

	b.registerCommands(b.commandFactory(nil), "")
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

		b.registerCommands(subCmd, toolName)
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

	// map for tool object
	flagMap := map[string]any{}

	// add local flags to flag map
	cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
		if flag.Hidden {
			return
		}
		b.logger.Debug("Registering Tool Parameter", slog.String("cmd", cmd.Name()), slog.String("name", flag.Name))
		// toolOptions = append(toolOptions, b.createParameterFromFlag(flag)...)
		flagMap[flag.Name] = flagToolOption(flag)
	})

	// add inherited flags to flag map
	cmd.InheritedFlags().VisitAll(func(flag *pflag.Flag) {
		if flag.Hidden {
			return
		}
		// Check if this flag was already added from local flags to avoid duplicates
		if cmd.Flags().Lookup(flag.Name) == nil {
			b.logger.Debug("Registering Tool Parameter", slog.String("cmd", cmd.Name()), slog.String("name", flag.Name))
			//toolOptions = append(toolOptions, b.createParameterFromFlag(flag)...)
			flagMap[flag.Name] = flagToolOption(flag)
		}
	})

	// add flags to tool
	toolOptions = append(toolOptions, mcp.WithObject("flags",
		mcp.Description("flag options"),
		mcp.Properties(flagMap),
	))

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
		result := b.executeCommand(ctx, cmd.CommandPath(), request)
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

// executeCommand executes the Cobra command using a fresh instance to avoid state pollution
func (b *CobraToMCPBridge) executeCommand(ctx context.Context, path string, request mcp.CallToolRequest) *mcp.CallToolResult {
	arguments := request.GetArguments()
	fields := strings.Fields(path)

	// get a new instance of the same cmd
	var output strings.Builder
	cmd := b.commandFactory(&output)

	// args must be added to root cmd
	var args []string
	if len(fields) > 1 {
		// Add command path to args
		args = append(args, fields[1:]...)
		b.logger.Debug("Set command arguments", "args", args)
	}
	// Handle positional arguments from the "args" parameter
	if argsValue, exists := arguments[PositionalArgsParam]; exists {
		if argsStr, ok := argsValue.(string); ok && argsStr != "" {
			// Split the args string by spaces to get individual arguments
			args = append(args, strings.Fields(argsStr)...)
			b.logger.Debug("Parsed positional arguments", "args", args)
		}
	}
	cmd.SetArgs(args)

	// flags must be set on relevant command
	if len(fields) > 1 {
		// move to subCommand
		for _, field := range fields {
			for _, subCmd := range cmd.Commands() {
				if field == subCmd.Name() {
					cmd = subCmd
					break
				}
			}
		}
		// verify cmd is set
		cmdPath := cmd.CommandPath()
		if cmdPath != path {
			b.logger.Error("command paths mismatch", slog.String("path", path), slog.String("cmdPath", cmdPath))
		}
	}

	// load flag map into cmd
	flags := arguments["flags"]
	b.logger.Debug("flags", "map", flags)
	flagMap, ok := flags.(map[string]any)
	if ok {
		for k, v := range flagMap {
			b.logger.Debug("setting flag", slog.String("cmd", cmd.Name()), slog.String("flag", k))
			flag := cmd.Flag(k)
			if flag == nil {
				b.logger.Error("flag not found", slog.String("cmd", cmd.Name()), slog.String("name", k))
				continue
			}
			err := flag.Value.Set(fmt.Sprintf("%v", v))
			if err != nil {
				b.logger.Error("Failed to load flag", slog.String("cmd", cmd.Name()), slog.String("key", k), slog.Any("value", v))
				return mcp.NewToolResultErrorFromErr(fmt.Sprintf("%s: failed to load flag %s", cmd.Name(), k), err)
			}
		}
	}

	// Execute the command's Run function with proper error recovery
	var result *mcp.CallToolResult
	func() {
		// Recover from any panics that might occur during command execution
		defer func() {
			if r := recover(); r != nil {
				b.logger.Error("Command execution panicked", "command", cmd.Name(), "panic", r)
				result = mcp.NewToolResultError("unexpected panic")
			}
		}()

		err := cmd.ExecuteContext(ctx)
		if err != nil {
			b.logger.Error("Command failed", "command", cmd.Name(), "error", err.Error())
			result = mcp.NewToolResultError(err.Error())
		} else {
			data := output.String()
			b.logger.Debug("Command succeeded", "command", cmd.Name(), "output", data)
			result = mcp.NewToolResultText(data)
		}
	}()

	return result
}

// StartServer starts the MCP server using stdio transport
func (b *CobraToMCPBridge) StartServer() error {
	return server.ServeStdio(b.server)
}
