package bridge

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

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

		// ignore mcp server commands
		if subCmd.Name() == MCPCommandName {
			continue
		}

		b.registerCommands(subCmd, toolName)
	}

	// Skip if the command has no runnable function
	if cmd.Run == nil && cmd.RunE == nil {
		return
	}

	b.logger.Debug("Registering command", "name", cmd.Name(), "path", parentPath)
	toolOptions := []mcp.ToolOption{
		mcp.WithDescription(b.getCommandDescription(cmd, parentPath)),
	}

	// add flags to tool
	flagMap := b.flagMapFromCmd(cmd)
	toolOptions = append(toolOptions, mcp.WithObject(FlagsParam,
		mcp.Description("flag options"),
		mcp.Properties(flagMap),
	))

	// Add an "args" parameter for positional arguments
	argsDescription := b.argsDescFromCmd(cmd)
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

func (b *CobraToMCPBridge) argsDescFromCmd(cmd *cobra.Command) string {
	argsDescription := "Space-separated positional arguments for the command"
	if cmd.Use != "" {
		argsDescription += fmt.Sprintf(". Usage: %s", cmd.Use)
	}
	if cmd.Args != nil {
		// Try to provide more specific information about expected arguments
		argsDescription += ". See command usage for argument requirements."
	}
	return argsDescription
}

func (b *CobraToMCPBridge) flagMapFromCmd(cmd *cobra.Command) map[string]any {
	// map for tool object
	flagMap := map[string]any{}
	// add local flags to flag map
	cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
		if flag.Hidden {
			return
		}

		b.logger.Debug("Registering Tool Parameter", slog.String("cmd", cmd.Name()), slog.String("name", flag.Name))
		flagMap[flag.Name] = flagToolOption(flag)
	})
	// add inherited flags to flag map
	cmd.InheritedFlags().VisitAll(func(flag *pflag.Flag) {
		if flag.Hidden {
			return
		}

		// Check if this flag was already added from local flags to avoid duplicates
		if _, ok := flagMap[flag.Name]; !ok {
			b.logger.Debug("Registering Tool Parameter", slog.String("cmd", cmd.Name()), slog.String("name", flag.Name))
			flagMap[flag.Name] = flagToolOption(flag)
		}
	})
	return flagMap
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

func flagToolOption(flag *pflag.Flag) map[string]string {
	description := flag.Usage
	if description == "" {
		description = fmt.Sprintf("Flag: %s", flag.Name)
	}

	return map[string]string{
		"type":        flag.Value.Type(),
		"description": description,
	}
}
