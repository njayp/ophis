# Ophis

**Transform any Cobra CLI application into an MCP (Model Context Protocol) server**

Ophis is a Go library that automatically converts Cobra-based command-line applications into MCP servers, allowing AI assistants and other MCP clients to interact with your CLI tools through structured protocols.

## What is MCP?

The [Model Context Protocol (MCP)](https://modelcontextprotocol.io/) is an open standard that enables secure connections between AI systems and external data sources and tools. By converting your CLI application to an MCP server, you make it accessible to AI assistants like Claude.

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

Ophis includes an example that exposes the entire Helm CLI as an MCP server:

```go
func createHelmCommand(output io.Writer) *cobra.Command {
    cmd, err := helmcmd.NewRootCmd(output, nil)
    if err != nil {
        panic(err)
    }
    return cmd
}

func main() {
    logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
    
    bridge := ophis.NewCobraToMCPBridge(
        createHelmCommand,
        "helm-mcp",
        "1.0.0",
        logger,
    )
    
    bridge.StartServer()
}
```

With this setup, an AI assistant can execute commands like `helm list`, `helm install myapp ./chart`, or `helm upgrade myapp ./chart --set key=value`.

## How It Works

1. **Command Discovery**: Recursively walks through your Cobra command tree
2. **Tool Registration**: Each command becomes an MCP tool with generated metadata
3. **Parameter Mapping**: Command flags are mapped to tool parameters
4. **Execution**: Creates a fresh command instance, sets flags/arguments, and executes
5. **Response**: Captures and returns command output to the MCP client

## Command Factory Pattern

Ophis uses a command factory to ensure clean state management:

```go
type CommandFactory func(output io.Writer) *cobra.Command
```

The factory creates a fresh command instance for each execution, preventing state pollution between tool calls.

## License

Licensed under the Apache License 2.0. See the [LICENSE](LICENSE) file for details.