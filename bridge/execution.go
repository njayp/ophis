package bridge

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/kballard/go-shellquote"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/njayp/ophis/tools"
	"github.com/spf13/cobra"
)

// executeCommand executes the Cobra command using a fresh instance to avoid state pollution.
// This method is safe for concurrent execution as it creates a new command instance
// for each request through the CommandFactory.New() method.
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
		b.logger.Error("Failed to find command in command tree",
			"error", err.Error(),
			"cmdPath", cmdPath)
		return mcp.NewToolResultError(fmt.Sprintf("Command not found: %s", err.Error()))
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

	// Add command path to args (skip the root command name)
	if len(cmdPath) > 1 {
		args = append(args, cmdPath[1:]...)
	}

	// Handle positional arguments from the "args" parameter
	if argsValue, ok := message[tools.PositionalArgsParam]; ok {
		if argsStr, ok := argsValue.(string); ok && argsStr != "" {
			// Use shell-like argument parsing to handle quoted strings properly
			parsedArgs := parseArgumentString(argsStr)
			args = append(args, parsedArgs...)

			// Log parsing details for debugging
			if len(parsedArgs) > 0 {
				b.logger.Debug("Parsed positional arguments",
					"raw", argsStr,
					"parsed", parsedArgs,
					"count", len(parsedArgs))
			}
		}
	}

	b.logger.Debug("Set command arguments", "args", args, "cmdPath", cmdPath)
	cmd.SetArgs(args)
}

// parseArgumentString provides shell-like argument parsing with proper quote handling.
// It supports single quotes, double quotes, and backslash escaping.
//
// The parsing is done using the github.com/kballard/go-shellquote library which
// follows /bin/sh word-splitting rules. This allows MCP clients to pass complex
// arguments containing spaces, quotes, and special characters.
//
// Examples:
//   - `foo bar baz` -> ["foo", "bar", "baz"]
//   - `foo "bar baz"` -> ["foo", "bar baz"]
//   - `foo 'bar baz'` -> ["foo", "bar baz"]
//   - `foo bar\ baz` -> ["foo", "bar baz"]
//
// If parsing fails due to malformed input (e.g., unterminated quotes), the function
// falls back to simple space-based splitting to ensure robustness.
func parseArgumentString(argsStr string) []string {
	// Trim whitespace and handle empty string
	argsStr = strings.TrimSpace(argsStr)
	if argsStr == "" {
		return nil
	}

	// Use shellquote to properly parse the arguments
	args, err := shellquote.Split(argsStr)
	if err != nil {
		// If parsing fails, fall back to simple splitting
		// This ensures we don't completely fail on malformed input
		return strings.Fields(argsStr)
	}

	return args
}

func (b *Manager) loadFlagsFromMap(cmd *cobra.Command, flagMap map[string]any) error {
	if cmd == nil {
		return fmt.Errorf("command cannot be nil")
	}
	if flagMap == nil {
		return nil // No flags to set
	}

	for k, v := range flagMap {
		// Validate flag name
		if k == "" {
			b.logger.Warn("Empty flag name provided")
			continue
		}

		b.logger.Debug("setting flag", slog.String("cmd", cmd.Name()), slog.String("flag", k))
		flag := cmd.Flag(k)
		if flag == nil {
			b.logger.Error("flag not found", slog.String("cmd", cmd.Name()), slog.String("name", k))
			continue
		}

		// Convert value to string with better handling
		var valueStr string
		if v == nil {
			valueStr = ""
		} else {
			valueStr = fmt.Sprintf("%v", v)
		}

		err := flag.Value.Set(valueStr)
		if err != nil {
			b.logger.Error("Failed to set flag", slog.String("cmd", cmd.Name()), slog.String("key", k), slog.Any("value", v), slog.String("error", err.Error()))
			return fmt.Errorf("%s: failed to set flag %s to value %v: %w", cmd.Name(), k, v, err)
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
