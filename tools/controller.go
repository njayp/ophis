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
	"github.com/njayp/ophis/internal/cfgmgr"
)

// Controller represents an MCP tool with its associated logic for execution and output handling.
type Controller struct {
	Tool    mcp.Tool `json:"tool"`
	handler Handler
}

// Handle processes the result of a tool execution into an MCP response.
func (c *Controller) Handle(ctx context.Context, request mcp.CallToolRequest, data []byte, err error) (*mcp.CallToolResult, error) {
	if c.handler != nil {
		// Use custom handler if provided
		return c.handler(ctx, request, data, err)
	}

	// Default handling: return output as plain text
	return DefaultHandler(ctx, request, data, err)
}

// Execute runs the tool command with the provided request.
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

// buildCommandArgs builds the command line arguments from the tool and request.
func (c *Controller) buildCommandArgs(request mcp.CallToolRequest) []string {
	message := request.GetArguments()

	// Start with the command path (e.g., "root_sub_command" -> ["root", "sub", "command"])
	// And remove the root command prefix
	args := strings.Split(c.Tool.Name, "_")[1:]
	slog.Debug("initial command arguments", "args", args)

	// Add flags
	if flagsValue, ok := message[cfgmgr.FlagsParam]; ok {
		if flagMap, ok := flagsValue.(map[string]any); ok {
			flagArgs := buildFlagArgs(flagMap)
			args = append(args, flagArgs...)
		}
	}

	// Add positional arguments
	if argsValue, ok := message[cfgmgr.PositionalArgsParam]; ok {
		if argsStr, ok := argsValue.(string); ok && argsStr != "" {
			parsedArgs := parseArgumentString(argsStr)
			args = append(args, parsedArgs...)
		}
	}

	return args
}

// buildFlagArgs converts a flag map to command line flag arguments.
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
		slog.Error("failed to parse argument string", "input", argsStr, "error", err)
		// If parsing fails, fall back to simple splitting
		// This ensures we don't completely fail on malformed input
		return strings.Fields(argsStr)
	}

	return args
}
