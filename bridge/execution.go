package bridge

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	sq "github.com/kballard/go-shellquote"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/njayp/ophis/tools"
)

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
	args, err := sq.Split(argsStr)
	if err != nil {
		// If parsing fails, fall back to simple splitting
		// This ensures we don't completely fail on malformed input
		return strings.Fields(argsStr)
	}

	return args
}

// executeCommand executes the command using exec.Cmd.
func (b *Manager) executeCommand(ctx context.Context, tool tools.Tool, request mcp.CallToolRequest) *mcp.CallToolResult {
	// Get the executable path
	executablePath, err := os.Executable()
	if err != nil {
		b.logger.Error("Failed to get executable path", "error", err.Error())
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get executable path: %s", err.Error()))
	}

	// Resolve any symlinks
	executablePath, err = filepath.EvalSymlinks(executablePath)
	if err != nil {
		b.logger.Error("Failed to resolve executable symlinks", "path", executablePath, "error", err.Error())
		return mcp.NewToolResultError(fmt.Sprintf("Failed to resolve executable symlinks: %s", err.Error()))
	}

	// Build command arguments
	cmdArgs, err := b.buildCommandArgs(tool, request)
	if err != nil {
		b.logger.Error("Failed to build command arguments", "error", err.Error())
		return mcp.NewToolResultError(fmt.Sprintf("Failed to build command arguments: %s", err.Error()))
	}

	// Create exec.Cmd
	cmd := exec.CommandContext(ctx, executablePath, cmdArgs...)

	b.logger.Info("Executing command via exec.Cmd",
		"executable", executablePath,
		"args", cmdArgs)

	// Execute the command
	data, err := cmd.CombinedOutput()
	output := string(data)

	if err != nil {
		b.logger.Error("Command execution failed",
			"error", err.Error(),
			"out", output,
		)

		// Include output in error message if available
		errMsg := fmt.Sprintf("Command execution failed: %s", err.Error())
		if output != "" {
			errMsg += fmt.Sprintf("\nOutput: %s", output)
		}
		return mcp.NewToolResultError(errMsg)
	}

	return mcp.NewToolResultText(output)
}

// buildCommandArgs builds the command line arguments from the tool and request.
func (b *Manager) buildCommandArgs(tool tools.Tool, request mcp.CallToolRequest) ([]string, error) {
	message := request.GetArguments()

	// Start with the command path (e.g., "root_sub_command" -> ["root", "sub", "command"])
	cmdPath := strings.Split(tool.Tool.Name, "_")
	args := make([]string, 0, len(cmdPath))

	// Add command path components
	args = append(args, cmdPath...)

	// Add flags
	if flagsValue, ok := message[tools.FlagsParam]; ok {
		if flagMap, ok := flagsValue.(map[string]any); ok {
			flagArgs, err := b.buildFlagArgs(flagMap)
			if err != nil {
				return nil, fmt.Errorf("failed to build flag arguments: %w", err)
			}
			args = append(args, flagArgs...)
		}
	}

	// Add positional arguments
	if argsValue, ok := message[tools.PositionalArgsParam]; ok {
		if argsStr, ok := argsValue.(string); ok && argsStr != "" {
			parsedArgs := parseArgumentString(argsStr)
			args = append(args, parsedArgs...)
		}
	}

	return args, nil
}

// buildFlagArgs converts a flag map to command line flag arguments.
func (b *Manager) buildFlagArgs(flagMap map[string]any) ([]string, error) {
	var args []string

	for name, value := range flagMap {
		if name == "" {
			continue
		}

		// Convert value to string
		valueStr := ""
		if value != nil {
			// Special handling for boolean flags
			if boolVal, ok := value.(bool); ok {
				if boolVal {
					// For true boolean flags, just add the flag name
					args = append(args, fmt.Sprintf("--%s", name))
				}
				// For false boolean flags, don't add anything
				continue
			}

			valueStr = fmt.Sprintf("%v", value)
		}

		// Add flag with value (for non-boolean flags)
		if valueStr != "" {
			args = append(args, fmt.Sprintf("--%s", name), valueStr)
		}
	}

	return args, nil
}
