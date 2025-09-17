package bridge

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"

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
	return append(args, input.Args...)
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
