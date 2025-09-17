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

var executablePath string

func initExecPath() {
	path, err := os.Executable()
	if err != nil {
		panic(fmt.Sprintf("Failed to get executable path: %v", err))
	}

	executablePath = path
}

// Execute runs the underlying CLI command.
func Execute(ctx context.Context, request *mcp.CallToolRequest, input *CmdToolInput) (*mcp.CallToolResult, *CmdToolOutput, error) {
	slog.Info("MCP tool request received", "request", request.Params.Name)
	// Build command arguments
	name := request.Params.Name
	args := buildCommandArgs(name, input)

	slog.Debug("executing command",
		"tool", name,
		"input", input,
		"args", args,
	)

	// Create exec.Cmd and run it
	cmd := exec.CommandContext(ctx, executablePath, args...)
	return execute(cmd)
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

	// if no input flags or args, return args
	if input == nil {
		return args
	}

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
