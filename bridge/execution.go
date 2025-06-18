package bridge

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/njayp/ophis/bridge/tools"
	"github.com/spf13/cobra"
)

// executeCommand executes the Cobra command using a fresh instance to avoid state pollution
func (b *Manager) executeCommand(ctx context.Context, tool tools.Tool, request mcp.CallToolRequest) *mcp.CallToolResult {
	message := request.GetArguments()
	cmdPath := strings.Split(tool.Tool.Name, "_")

	// get a new instance of the same cmd
	cmd, exec := b.commandFactory.New()

	// args must be added to root cmd
	b.loadArgs(cmd, cmdPath, message)

	// descendCmdTree to called command for flag access
	cmd, err := descendCmdTree(cmd, cmdPath)
	if err != nil {
		b.logger.Error(err.Error())
		return mcp.NewToolResultError(err.Error())
	}

	// load flag map into cmd
	flags := message[tools.FlagsParam]
	b.logger.Debug("flags", "map", flags)
	flagMap, ok := flags.(map[string]any)
	if ok {
		err := b.loadFlagsFromMap(cmd, flagMap)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("failed to load flags", err)
		}
	}

	// Check if context is already cancelled
	if err := ctx.Err(); err != nil {
		b.logger.Warn("Context cancelled before execution", "error", err)
		return mcp.NewToolResultError("request cancelled")
	}

	// Execute the command's Run function with proper error recovery
	var result *mcp.CallToolResult
	func() {
		// Recover from any panics that might occur during command execution
		defer func() {
			if r := recover(); r != nil {
				b.logger.Error("Command execution panicked", "command", cmd.Name(), "panic", r)
				result = mcp.NewToolResultError("command execution failed due to unexpected error")
			}
		}()

		result = exec(ctx)
	}()

	return result
}

func (b *Manager) loadArgs(cmd *cobra.Command, cmdPath []string, message map[string]any) {
	var args []string

	// Add command path to args
	if len(cmdPath) > 1 {
		args = append(args, cmdPath[1:]...)
	}

	// Handle positional arguments from the "args" parameter
	if argsValue, ok := message[tools.PositionalArgsParam]; ok {
		if argsStr, ok := argsValue.(string); ok && argsStr != "" {
			// Split the args string by spaces to get individual arguments
			args = append(args, strings.Fields(argsStr)...)
		}
	}

	b.logger.Debug("Set command arguments", "args", args)
	cmd.SetArgs(args)
}

func (b *Manager) loadFlagsFromMap(cmd *cobra.Command, flagMap map[string]any) error {
	if cmd == nil {
		return fmt.Errorf("command cannot be nil")
	}
	if flagMap == nil {
		return nil // No flags to set
	}

	for k, v := range flagMap {
		b.logger.Debug("setting flag", slog.String("cmd", cmd.Name()), slog.String("flag", k))
		flag := cmd.Flag(k)
		if flag == nil {
			b.logger.Error("flag not found", slog.String("cmd", cmd.Name()), slog.String("name", k))
			continue
		}
		err := flag.Value.Set(fmt.Sprintf("%v", v))
		if err != nil {
			b.logger.Error("Failed to set flag", slog.String("cmd", cmd.Name()), slog.String("key", k), slog.Any("value", v))
			return fmt.Errorf("%s: failed to set flag %s", cmd.Name(), k)
		}
	}

	return nil
}

func descendCmdTree(cmd *cobra.Command, cmdPath []string) (*cobra.Command, error) {
	// flags must be set on relevant command
	if len(cmdPath) > 1 {
		// move to subCommand
		for _, field := range cmdPath {
			for _, subCmd := range cmd.Commands() {
				if field == subCmd.Name() {
					cmd = subCmd
					break
				}
			}
		}
	}

	// verify cmd is set
	newPath := cmd.CommandPath()
	if newPath != strings.Join(cmdPath, " ") {
		return nil, fmt.Errorf("command path not recognized: %s", cmdPath)
	}
	return cmd, nil
}
