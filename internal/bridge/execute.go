package bridge

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"

	sq "github.com/kballard/go-shellquote"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Execute runs the underlying CLI command.
func Execute() mcp.ToolHandlerFor[*CmdToolInput, *CmdToolOutput] {
	return func(ctx context.Context, request *mcp.CallToolRequest, input *CmdToolInput) (result *mcp.CallToolResult, output *CmdToolOutput, _ error) {
		slog.Info("MCP tool request received", "request", request.Params.Name)
		// Get the executable path
		executablePath, err := os.Executable()
		if err != nil {
			slog.Error("failed to get executable path", "error", err)
			return nil, nil, fmt.Errorf("failed to get executable path: %w", err)
		}

		// Build command arguments
		name := request.Params.Name
		cmdArgs := buildCommandArgs(name, input)

		slog.Debug("executing command",
			"tool", name,
			"executable", executablePath,
			"args", cmdArgs,
		)

		// Create exec.Cmd and run it
		cmd := exec.CommandContext(ctx, executablePath, cmdArgs...)
		return execute(cmd)
	}
}

// execute runs the given exec.Cmd and returns stdout, stderr, and exit code.
func execute(cmd *exec.Cmd) (*mcp.CallToolResult, *CmdToolOutput, error) {
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	exitCode := 0
	if err != nil {
		// Check if it's an ExitError to get the exit code
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			// Non-exit errors (like command not found)
			return nil, &CmdToolOutput{
				StdOut:   stdout.String(),
				StdErr:   stderr.String(),
				ExitCode: -1,
			}, err
		}
	} else {
		// Successful run
		if cmd.ProcessState != nil {
			exitCode = cmd.ProcessState.ExitCode()
		}
	}

	return nil, &CmdToolOutput{
		StdOut:   stdout.String(),
		StdErr:   stderr.String(),
		ExitCode: exitCode,
	}, nil
}

// buildCommandArgs constructs CLI arguments from the MCP request.
func buildCommandArgs(name string, input *CmdToolInput) []string {
	// Start with the command path (e.g., "root_sub_command" -> ["root", "sub", "command"])
	// And remove the root command prefix
	args := strings.Split(name, "_")[1:]
	slog.Debug("initial command arguments", "args", args)

	// Add flags
	flagArgs := buildFlagArgs(input.Flags)
	args = append(args, flagArgs...)

	// Add positional arguments
	parsedArgs := parseArgumentString(input.Args)
	return append(args, parsedArgs...)
}

// buildFlagArgs converts MCP flags to CLI flag arguments.
func buildFlagArgs(flagMap map[string]any) []string {
	var args []string

	for name, value := range flagMap {
		if name == "" || value == nil {
			continue
		}

		if items, ok := value.([]any); ok {
			for _, item := range items {
				slog.Debug("adding flag slice argument", "flag_name", name, "input", value, "value", item)
				args = append(args, parseFlagArgValue(name, item)...)
			}

			continue
		}

		args = append(args, parseFlagArgValue(name, value)...)
	}

	return args
}

func parseFlagArgValue(name string, value any) (retVal []string) {
	if value != nil {
		switch v := value.(type) {
		case bool:
			if v {
				slog.Debug("adding boolean flag argument", "flag_name", name, "value", v)
				retVal = append(retVal, fmt.Sprintf("--%s", name))
			}
		default:
			slog.Debug("adding flag argument", "flag_name", name, "value", value)
			retVal = append(retVal, fmt.Sprintf("--%s", name), fmt.Sprintf("%v", value))
		}
	}

	return retVal
}

// parseArgumentString parses shell-like arguments with quote handling.
// Supports single quotes, double quotes, and backslash escaping.
// Falls back to space splitting on parse errors.
func parseArgumentString(argsStr string) []string {
	// Trim whitespace and handle empty string
	argsStr = strings.TrimSpace(argsStr)
	if argsStr == "" {
		return nil
	}

	// Use shellquote to properly parse the arguments
	args, err := sq.Split(argsStr)
	if err != nil {
		slog.Error("failed to parse argument string", "input", argsStr, "error", err)
		// If parsing fails, fall back to simple splitting
		// This ensures we don't completely fail on malformed input
		return strings.Fields(argsStr)
	}

	return args
}
