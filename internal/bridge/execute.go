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

var executablePath = initExecPath()

func initExecPath() string {
	path, err := os.Executable()
	if err != nil {
		panic(fmt.Sprintf("Failed to get executable path: %v", err))
	}

	return path
}

func (s *Selector) execute(ctx context.Context, request *mcp.CallToolRequest, input ToolInput) (result *mcp.CallToolResult, output ToolOutput, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()

	if s.PreRun != nil {
		ctx, request, input = s.PreRun(ctx, request, input)
	}

	result, output, err = execute(ctx, request, input)

	if s.PostRun != nil {
		result, output, err = s.PostRun(ctx, request, input, result, output, err)
	}

	return result, output, err
}

// execute runs the underlying CLI command.
func execute(ctx context.Context, request *mcp.CallToolRequest, input ToolInput) (*mcp.CallToolResult, ToolOutput, error) {
	name := request.Params.Name
	slog.Info("mcp tool request received", "request", name)

	// Build command arguments
	args := buildCommandArgs(name, input)
	slog.Debug("executing command",
		"tool", name,
		"input", input,
		"args", args,
	)

	// Create exec.Cmd and run it
	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, executablePath, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	exitCode := 0

	err := cmd.Run()
	if err != nil {
		// Check if it's an ExitError to get the exit code
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			// Non-exit errors (like command not found)
			slog.Error("command failed to run", "name", name, "error", err)
			return nil, ToolOutput{}, err
		}
	}

	return nil, ToolOutput{
		StdOut:   stdout.String(),
		StdErr:   stderr.String(),
		ExitCode: exitCode,
	}, nil
}

// buildCommandArgs constructs CLI arguments from the MCP request.
func buildCommandArgs(name string, input ToolInput) []string {
	// Start with the command path (e.g., "root_sub_command" -> ["root", "sub", "command"])
	// And remove the root command prefix
	args := strings.Split(name, "_")[1:]

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
				retVal = append(retVal, fmt.Sprintf("--%s", name))
			}
		default:
			retVal = append(retVal, fmt.Sprintf("--%s", name), fmt.Sprintf("%v", value))
		}
	}

	return retVal
}
