package ophis

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/spf13/cobra"
)

// executeCommand executes the Cobra command using a fresh instance to avoid state pollution
func (b *CobraToMCPBridge) executeCommand(ctx context.Context, cmdPath string, request mcp.CallToolRequest) *mcp.CallToolResult {
	message := request.GetArguments()

	// get a new instance of the same cmd
	cmd, exec := b.commandFactory.CreateCommand()

	// args must be added to root cmd
	b.loadArgs(cmd, cmdPath, message)

	// descendCmdTree to called command for flag access
	cmd, err := descendCmdTree(cmd, cmdPath)
	if err != nil {
		b.logger.Error(err.Error())
		return mcp.NewToolResultError(err.Error())
	}

	// load flag map into cmd
	flags := message[FlagsParam]
	b.logger.Debug("flags", "map", flags)
	flagMap, ok := flags.(map[string]any)
	if ok {
		err := b.loadFlagsFromMap(cmd, flagMap)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("failed to load flags", err)
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

		result = exec(ctx)
	}()

	return result
}

func (b *CobraToMCPBridge) loadArgs(cmd *cobra.Command, cmdPath string, message map[string]any) {
	fields := strings.Fields(cmdPath)
	var args []string
	if len(fields) > 1 {
		// Add command path to args
		args = append(args, fields[1:]...)
		b.logger.Debug("Set command arguments", "args", args)
	}
	// Handle positional arguments from the "args" parameter
	if argsValue, ok := message[PositionalArgsParam]; ok {
		if argsStr, ok := argsValue.(string); ok && argsStr != "" {
			// Split the args string by spaces to get individual arguments
			args = append(args, strings.Fields(argsStr)...)
			b.logger.Debug("Parsed positional arguments", "args", args)
		}
	}
	cmd.SetArgs(args)
}

func (b *CobraToMCPBridge) loadFlagsFromMap(cmd *cobra.Command, flagMap map[string]any) error {
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
