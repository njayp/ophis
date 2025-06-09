package ophis

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// CobraToMCPBridge converts a Cobra CLI application to an MCP server
type CobraToMCPBridge struct {
	rootCmd        *cobra.Command
	appName        string
	version        string
	server         *server.MCPServer
	executablePath string
}

// NewCobraToMCPBridge creates a new bridge instance
func NewCobraToMCPBridge(rootCmd *cobra.Command, appName, version string) *CobraToMCPBridge {
	// Get the executable path for the current binary
	executablePath, err := os.Executable()
	if err != nil {
		executablePath = os.Args[0] // fallback to program name
	}

	return &CobraToMCPBridge{
		rootCmd:        rootCmd,
		appName:        appName,
		version:        version,
		executablePath: executablePath,
	}
}

// SetExecutablePath allows overriding the executable path (useful for testing or custom setups)
func (b *CobraToMCPBridge) SetExecutablePath(path string) {
	b.executablePath = path
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
	// Skip the root command if it has no Run function and has subcommands
	if parentPath == "" && cmd.Run == nil && cmd.RunE == nil && len(cmd.Commands()) > 0 {
		// Register subcommands
		for _, subCmd := range cmd.Commands() {
			if subCmd.Hidden {
				continue
			}
			b.registerCommands(subCmd, subCmd.Name())
		}
		return
	}

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

	// Add parameters for flags
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		if flag.Hidden {
			return
		}
		toolOptions = append(toolOptions, b.createParameterFromFlag(flag)...)
	})

	/*
		// Add positional arguments parameter if the command accepts them
		if cmd.Args == nil || cmd.Args == cobra.ArbitraryArgs || cmd.Args == cobra.MinimumNArgs(0) {
			toolOptions = append(toolOptions, mcp.WithArray("args",
				mcp.Description("Positional arguments for the command"),
				mcp.WithItems(mcp.NewSchemaBuilder().String().Build()),
			))
		}
	*/

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
	/*
		case "stringSlice", "stringArray":
			options = append(options, mcp.WithArray(flagName,
				mcp.Description(description),
				mcp.WithItems(mcp.NewSchemaBuilder().String().Build()),
			))
		case "intSlice":
			options = append(options, mcp.WithArray(flagName,
				mcp.Description(description),
				mcp.WithItems(mcp.NewSchemaBuilder().Number().Build()),
			))
	*/
	default:
		// Default to string for unknown types
		options = append(options, mcp.WithString(flagName,
			mcp.Description(description),
		))
	}

	return options
}

// executeCommand executes the Cobra command with the provided parameters
func (b *CobraToMCPBridge) executeCommand(ctx context.Context, cmd *cobra.Command, parentPath string, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Build command arguments
	args := []string{}

	// Add command path
	if parentPath != "" {
		parts := strings.Split(parentPath, "_")
		args = append(args, parts...)
	}
	args = append(args, cmd.Name())

	// Process flags
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		if flag.Hidden {
			return
		}

		flagName := strings.ReplaceAll(flag.Name, "-", "_")
		value, exists := request.Params.Arguments[flagName]
		if !exists {
			return
		}

		flagArg := "--" + flag.Name

		switch flag.Value.Type() {
		case "bool":
			if boolVal, ok := value.(bool); ok && boolVal {
				args = append(args, flagArg)
			}
		case "string":
			if strVal, ok := value.(string); ok && strVal != "" {
				args = append(args, flagArg, strVal)
			}
		case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
			if numVal, ok := value.(float64); ok {
				args = append(args, flagArg, strconv.FormatInt(int64(numVal), 10))
			}
		case "float32", "float64":
			if numVal, ok := value.(float64); ok {
				args = append(args, flagArg, strconv.FormatFloat(numVal, 'f', -1, 64))
			}
		case "stringSlice", "stringArray":
			if arrayVal, ok := value.([]interface{}); ok {
				for _, item := range arrayVal {
					if strItem, ok := item.(string); ok {
						args = append(args, flagArg, strItem)
					}
				}
			}
		case "intSlice":
			if arrayVal, ok := value.([]interface{}); ok {
				for _, item := range arrayVal {
					if numItem, ok := item.(float64); ok {
						args = append(args, flagArg, strconv.FormatInt(int64(numItem), 10))
					}
				}
			}
		default:
			// Try to convert to string
			if strVal := fmt.Sprintf("%v", value); strVal != "" && strVal != "<nil>" {
				args = append(args, flagArg, strVal)
			}
		}
	})

	// Add positional arguments
	if argsVal, exists := request.Params.Arguments["args"]; exists {
		if argsArray, ok := argsVal.([]interface{}); ok {
			for _, arg := range argsArray {
				if strArg, ok := arg.(string); ok {
					args = append(args, strArg)
				}
			}
		}
	}

	// Execute the command
	execCmd := exec.CommandContext(ctx, b.executablePath, args...)
	output, err := execCmd.CombinedOutput()
	if err != nil {
		slog.Error("Failed to execute command", "command", execCmd.String(), "error", err, "output", string(output))

		// Check if it's an exit code error
		if exitError, ok := err.(*exec.ExitError); ok {
			return mcp.NewToolResultError(fmt.Sprintf("Command failed with exit code %d:\n%s",
				exitError.ExitCode(), string(output))), nil
		}
		return mcp.NewToolResultError(fmt.Sprintf("Failed to execute command: %v\nOutput: %s",
			err, string(output))), nil
	}

	slog.Debug("Command executed successfully", "command", execCmd.String(), "output", string(output))
	return mcp.NewToolResultText(string(output)), nil
}

// StartServer starts the MCP server using stdio transport
func (b *CobraToMCPBridge) StartServer() error {
	if b.server == nil {
		b.CreateMCPServer()
	}
	return server.ServeStdio(b.server)
}

// GetServer returns the underlying MCP server (useful for custom transport)
func (b *CobraToMCPBridge) GetServer() *server.MCPServer {
	if b.server == nil {
		b.CreateMCPServer()
	}
	return b.server
}
