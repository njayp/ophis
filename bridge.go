package ophis

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// CobraToMCPBridge converts a Cobra CLI application to an MCP server
type CobraToMCPBridge struct {
	rootCmd *cobra.Command
	appName string
	version string
	server  *server.MCPServer
}

// NewCobraToMCPBridge creates a new bridge instance
func NewCobraToMCPBridge(rootCmd *cobra.Command, appName, version string) *CobraToMCPBridge {
	return &CobraToMCPBridge{
		rootCmd: rootCmd,
		appName: appName,
		version: version,
	}
}

// CreateMCPServer creates and configures the MCP server with tools for each Cobra command
func (b *CobraToMCPBridge) CreateMCPServer() *server.MCPServer {
	slog.Info("Creating MCP server", "app_name", b.appName, "version", b.version)

	s := server.NewMCPServer(
		b.appName,
		b.version,
	)

	b.server = s
	slog.Info("Registering Cobra commands as MCP tools")
	b.registerCommands(b.rootCmd, "")
	slog.Info("MCP server created successfully")
	return s
}

// registerCommands recursively registers all Cobra commands as MCP tools
func (b *CobraToMCPBridge) registerCommands(cmd *cobra.Command, parentPath string) {
	// Create the tool name
	toolName := cmd.Name()
	if parentPath != "" {
		toolName = parentPath + "_" + cmd.Name()
	}

	// Skip if the command has no runnable function
	if cmd.Run == nil && cmd.RunE == nil {
		// Still register subcommands
		for _, subCmd := range cmd.Commands() {
			if subCmd.Hidden {
				continue
			}
			subPath := toolName
			if parentPath == "" {
				subPath = cmd.Name()
			}
			b.registerCommands(subCmd, subPath)
		}
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
	toolOptions = append(toolOptions, mcp.WithString("args",
		mcp.Description(argsDescription),
	))

	// Create the tool
	tool := mcp.NewTool(toolName, toolOptions...)
	slog.Debug("Registering MCP tool", "tool_name", toolName, "command", cmd.Name())

	// Add the tool handler
	b.server.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		slog.Info("MCP tool request received", "tool_name", toolName, "arguments", request.Params.Arguments)
		result, err := b.executeCommand(ctx, cmd, parentPath, request)
		if err != nil {
			slog.Error("MCP tool execution failed", "tool_name", toolName, "error", err)
		} else {
			slog.Info("MCP tool executed successfully", "tool_name", toolName)
		}
		return result, err
	})

	// Register subcommands
	for _, subCmd := range cmd.Commands() {
		if subCmd.Hidden {
			continue
		}
		subPath := toolName
		if parentPath == "" && cmd.Name() != toolName {
			subPath = cmd.Name()
		}
		slog.Debug("Registering subcommand", "name", subCmd.Name(), "path", subPath)
		b.registerCommands(subCmd, subPath)
	}
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

	slog.Debug("Command description", "name", cmd.Name(), "description", desc)
	return desc
}

// createParameterFromFlag converts a Cobra flag to MCP tool parameters
func (b *CobraToMCPBridge) createParameterFromFlag(flag *pflag.Flag) []mcp.ToolOption {
	var options []mcp.ToolOption
	flagName := strings.ReplaceAll(flag.Name, "-", "_")

	description := flag.Usage
	if description == "" {
		description = fmt.Sprintf("Flag: %s", flag.Name)
	}

	// Determine parameter type based on flag type
	switch flag.Value.Type() {
	case "bool":
		options = append(options, mcp.WithBoolean(flagName,
			mcp.Description(description),
		))
	case "int", "int8", "int16", "int32", "int64":
		options = append(options, mcp.WithNumber(flagName,
			mcp.Description(description),
		))
	case "uint", "uint8", "uint16", "uint32", "uint64":
		options = append(options, mcp.WithNumber(flagName,
			mcp.Description(description),
		))
	case "float32", "float64":
		options = append(options, mcp.WithNumber(flagName,
			mcp.Description(description),
		))
	case "string":
		options = append(options, mcp.WithString(flagName,
			mcp.Description(description),
		))
	default:
		// Default to string for unknown types
		options = append(options, mcp.WithString(flagName,
			mcp.Description(description),
		))
	}

	return options
}

// executeCommand executes the Cobra command directly by calling its Run function
func (b *CobraToMCPBridge) executeCommand(ctx context.Context, cmd *cobra.Command, parentPath string, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Process arguments and set flags DIRECTLY on the original command
	// We'll reset flags after execution to avoid state pollution

	arguments := request.GetArguments()
	var args []string

	// Handle positional arguments from the "args" parameter
	if argsValue, exists := arguments["args"]; exists {
		if argsStr, ok := argsValue.(string); ok && argsStr != "" {
			// Split the args string by spaces to get individual arguments
			args = strings.Fields(argsStr)
			slog.Debug("Parsed positional arguments", "args", args)
		}
	}

	// Store original flag values for restoration later
	originalValues := make(map[string]string)

	// Debug: log all available flags
	slog.Debug("Available flags on command", "command", cmd.Name())
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		slog.Debug("Local flag", "name", flag.Name, "type", flag.Value.Type(), "value", flag.Value.String())
	})
	cmd.PersistentFlags().VisitAll(func(flag *pflag.Flag) {
		slog.Debug("Persistent flag", "name", flag.Name, "type", flag.Value.Type(), "value", flag.Value.String())
	})

	// Check parent commands for persistent flags
	if cmd.Parent() != nil {
		slog.Debug("Checking parent command flags", "parent", cmd.Parent().Name())
		cmd.Parent().PersistentFlags().VisitAll(func(flag *pflag.Flag) {
			slog.Debug("Parent persistent flag", "name", flag.Name, "type", flag.Value.Type(), "value", flag.Value.String())
		})
	}

	// Helper function to set flag values directly on the correct command
	setFlagValue := func(flag *pflag.Flag, targetCmd *cobra.Command) {
		if flag.Hidden {
			return
		}

		flagName := strings.ReplaceAll(flag.Name, "-", "_")
		// Skip the special "args" parameter as it's not a flag
		if flagName == "args" {
			return
		}
		value, exists := arguments[flagName]
		if !exists {
			return
		}

		// Store original value for restoration
		originalValues[flag.Name] = flag.Value.String()

		slog.Debug("Setting flag", "flag_name", flag.Name, "mcp_name", flagName, "value", value, "original", originalValues[flag.Name], "target_cmd", targetCmd.Name())

		// Set the flag value on the target command
		var err error
		switch flag.Value.Type() {
		case "bool":
			if boolVal, ok := value.(bool); ok {
				err = targetCmd.Flags().Set(flag.Name, strconv.FormatBool(boolVal))
				if err != nil {
					slog.Error("Failed to set bool flag", "flag", flag.Name, "value", boolVal, "error", err)
				} else {
					slog.Debug("Successfully set bool flag", "flag", flag.Name, "value", boolVal)
				}
			}
		case "string":
			if strVal, ok := value.(string); ok {
				err = targetCmd.Flags().Set(flag.Name, strVal)
				if err != nil {
					slog.Error("Failed to set string flag", "flag", flag.Name, "value", strVal, "error", err)
				} else {
					slog.Debug("Successfully set string flag", "flag", flag.Name, "value", strVal)
				}
			}
		case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
			if numVal, ok := value.(float64); ok {
				err = targetCmd.Flags().Set(flag.Name, strconv.FormatInt(int64(numVal), 10))
				if err != nil {
					slog.Error("Failed to set numeric flag", "flag", flag.Name, "value", numVal, "error", err)
				}
			}
		case "float32", "float64":
			if numVal, ok := value.(float64); ok {
				err = targetCmd.Flags().Set(flag.Name, strconv.FormatFloat(numVal, 'f', -1, 64))
				if err != nil {
					slog.Error("Failed to set float flag", "flag", flag.Name, "value", numVal, "error", err)
				}
			}
		case "stringArray", "stringSlice", "intArray", "intSlice", "uintArray", "uintSlice":
			slog.Error(("stringArray type is not supported in MCP tools"))
		default:
			// Try to convert to string
			if strVal := fmt.Sprintf("%v", value); strVal != "" && strVal != "<nil>" {
				err = targetCmd.Flags().Set(flag.Name, strVal)
				if err != nil {
					slog.Error("Failed to set default flag", "flag", flag.Name, "value", strVal, "error", err)
				}
			}
		}
	}

	// Process flags from MCP arguments and set them directly on the original command
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		setFlagValue(flag, cmd)
	})

	// Process persistent flags
	cmd.PersistentFlags().VisitAll(func(flag *pflag.Flag) {
		setFlagValue(flag, cmd)
	})

	// Capture output by redirecting the command's output
	var output strings.Builder
	cmd.SetOut(&output)
	cmd.SetErr(&output)

	// Execute the command's Run function directly
	var err error
	if cmd.RunE != nil {
		err = cmd.RunE(cmd, args)
	} else if cmd.Run != nil {
		cmd.Run(cmd, args)
	} else {
		return mcp.NewToolResultError("Command has no Run or RunE function"), nil
	}

	if err != nil {
		slog.Error("Failed to execute command function", "command", cmd.Name(), "error", err, "output", output.String())
		return mcp.NewToolResultError(fmt.Sprintf("Command failed: %v\nOutput: %s", err, output.String())), nil
	}

	result := output.String()
	slog.Debug("Command executed successfully", "command", cmd.Name(), "output", result)
	return mcp.NewToolResultText(result), nil
}

// StartServer starts the MCP server using stdio transport
func (b *CobraToMCPBridge) StartServer() error {
	if b.server == nil {
		b.CreateMCPServer()
	}
	return server.ServeStdio(b.server)
}
