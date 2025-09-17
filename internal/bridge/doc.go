// Package bridge provides utilities for creating MCP (Model Context Protocol) tools
// from Cobra commands with automatic schema generation.
//
// This package simplifies the process of converting Cobra CLI commands into MCP tools
// that can be used with Claude and other MCP clients. It automatically generates
// JSON schemas for command flags and arguments, and provides type-safe handlers.
//
// # Basic Usage
//
// The simplest way to create a tool from a command:
//
//	tool, handler := bridge.CreateToolFromCmd(myCommand, bridge.DefaultCommandHandler)
//
// # Custom Handler
//
// For actual command execution, provide a custom handler:
//
//	myHandler := func(ctx context.Context, req *mcp.CallToolRequest, input bridge.CmdToolInput) (*mcp.CallToolResult, error) {
//		// Execute the actual command with input.Flags and input.Args
//		output := executeCommand(input.Flags, input.Args)
//
//		return &mcp.CallToolResult{
//			Content: []mcp.Content{
//				&mcp.TextContent{Text: output},
//			},
//		}, nil
//	}
//
//	tool, handler := bridge.CreateToolFromCmd(myCommand, myHandler)
//
// # Working with Multiple Commands
//
// Create tools from multiple commands with filtering:
//
//	commands := []*cobra.Command{cmd1, cmd2, cmd3}
//
//	// Filter commands
//	filtered := bridge.FilterCommands(commands,
//		bridge.ExcludeHidden,
//		bridge.HasRunFunc,
//		bridge.ExcludeByName("help", "completion"),
//	)
//
//	// Create tools
//	tools := bridge.CreateToolsFromCommands(filtered, myHandler)
//
// # Input Structure
//
// Tools created by this package expect input in the following format:
//
//	{
//		"flags": {
//			"flag-name": "value",
//			"boolean-flag": true,
//			"array-flag": ["item1", "item2"]
//		},
//		"args": "positional arguments as a string"
//	}
//
// # Output Structure
//
// Tools return output in the following format:
//
//	{
//		"output": "command execution output",
//		"exitCode": 0,
//		"error": "error message if any"
//	}
//
// # Schema Generation
//
// The package automatically generates JSON schemas for:
// - Command flags with appropriate types (string, boolean, integer, array, etc.)
// - Positional arguments with usage patterns extracted from cmd.Use
// - Comprehensive descriptions built from cmd.Long, cmd.Short, and cmd.Example
//
// # Compatibility
//
// This package works with:
// - github.com/modelcontextprotocol/go-sdk/mcp for MCP types
// - github.com/google/jsonschema-go/jsonschema for schema generation
// - github.com/spf13/cobra for command introspection
//
// # Design Principles
//
// This package prioritizes:
// - Simplicity over configuration
// - Type safety through generated schemas
// - Automatic discovery of command metadata
// - Clean separation between schema generation and command execution
package bridge
