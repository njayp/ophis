package tools

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"

	sq "github.com/kballard/go-shellquote"
	"github.com/mark3labs/mcp-go/mcp"
)

// Controller executes a Cobra command as an MCP tool.
type Controller struct {
	Tool    mcp.Tool `json:"tool"`
	handler Handler
}

// Handle processes execution results using the configured handler.
func (c *Controller) Handle(ctx context.Context, request mcp.CallToolRequest, data []byte, err error) (*mcp.CallToolResult, error) {
	if c.handler != nil {
		// Use custom handler if provided
		return c.handler(ctx, request, data, err)
	}

	// Default handling: return output as plain text
	return DefaultHandler(ctx, request, data, err)
}

// Execute runs the underlying CLI command.
func (c *Controller) Execute(ctx context.Context, request mcp.CallToolRequest) ([]byte, error) {
	// Get the executable path
	executablePath, err := os.Executable()
	if err != nil {
		slog.Error("failed to get executable path", "error", err)
		return nil, fmt.Errorf("failed to get executable path: %w", err)
	}

	// Build command arguments
	cmdArgs := c.buildCommandArgs(request)

	slog.Debug("executing command",
		"tool", c.Tool.Name,
		"executable", executablePath,
		"args", cmdArgs,
	)

	// Create exec.Cmd and run it
	cmd := exec.CommandContext(ctx, executablePath, cmdArgs...)
	return cmd.CombinedOutput()
}

// buildCommandArgs constructs CLI arguments from the MCP request.
func (c *Controller) buildCommandArgs(request mcp.CallToolRequest) []string {
	message := request.GetArguments()

	// Start with the command path (e.g., "root_sub_command" -> ["root", "sub", "command"])
	// And remove the root command prefix
	args := strings.Split(c.Tool.Name, "_")[1:]
	slog.Debug("initial command arguments", "args", args)

	// Add flags
	if flagsValue, ok := message[flagsParam]; ok {
		if flagMap, ok := flagsValue.(map[string]any); ok {
			flagArgs := buildFlagArgs(flagMap)
			args = append(args, flagArgs...)
		}
	}

	// Add positional arguments
	if argsValue, ok := message[argsParam]; ok {
		if argsStr, ok := argsValue.(string); ok && argsStr != "" {
			parsedArgs := parseArgumentString(argsStr)
			args = append(args, parsedArgs...)
		}
	}

	return args
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
