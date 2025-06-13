# Ophis

**Transform any Cobra CLI application into an MCP (Model Context Protocol) server**

Ophis is a Go library that automatically converts Cobra-based command-line applications into MCP servers, allowing AI assistants and other MCP clients to interact with your CLI tools through structured protocols.

## Motivation

- Transform your cobra CLI to an MCP server with **one command**.
- Expose CLIs without exposing shell.

## Features

- **Automatic Tool Registration**: Recursively discovers and registers all Cobra commands as MCP tools
- **Flag Support**: Maps Cobra flags to MCP tool parameters with proper type detection
- **Positional Arguments**: Handles command-line arguments through a dedicated parameter
- **Command Hierarchy**: Preserves the structure of nested subcommands
- **Clean State Management**: Each command execution uses a fresh command instance

## Installation

```bash
go get github.com/ophis
```

## Helm Example

Ophis includes an example that exposes the entire Helm CLI as an MCP server. With this setup, an AI assistant can execute commands like `helm list`, `helm install myapp ./chart`, or `helm upgrade myapp ./chart --set key=value`.

## How It Works

1. **Command Discovery**: Recursively walks through your Cobra command tree
2. **Tool Registration**: Each command becomes an MCP tool with generated metadata
3. **Parameter Mapping**: Command flags are mapped to tool parameters
4. **Execution**: Creates a fresh command instance, sets flags/arguments, and executes
5. **Response**: Captures and returns command output to the MCP client

## Command Factory Pattern

Provide ophis with a command factory for flexible cmd creation and execution:

```go
type CommandFactory interface {
	CreateRegistrationCommand() *cobra.Command
	CreateCommand() (*cobra.Command, func(context.Context) *mcp.CallToolResult)
}
```

The factory creates a fresh command instance for each execution, preventing state pollution between tool calls.

## License

Licensed under the Apache License 2.0. See the [LICENSE](LICENSE) file for details.