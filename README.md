![Project Logo](./logo.png)

**Transform any Cobra CLI into an MCP server**

Ophis automatically converts your Cobra commands into MCP tools, and provides CLI commands for integration with Claude Desktop, VSCode, and Cursor.

## Quick Start

### Install

```bash
go get github.com/njayp/ophis
```

### Add to your CLI

```go
package main

import (
    "os"
    "github.com/njayp/ophis"
)

func main() {
    rootCmd := createMyRootCommand()
    rootCmd.AddCommand(ophis.Command(nil))

    if err := rootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}
```

### Enable in Claude Desktop, VSCode, or Cursor

```bash
# Claude Desktop
./my-cli mcp claude enable
# Restart Claude Desktop

# VSCode (requires Copilot in Agent Mode)
./my-cli mcp vscode enable

# Cursor
./my-cli mcp cursor enable
```

Your CLI commands are now available as MCP tools!

### Stream over HTTP

Expose your MCP server over HTTP for remote access:

```bash
./my-cli mcp stream --host localhost --port 8080
```

## Commands

The `ophis.Command(nil)` adds these subcommands to your CLI:

```
mcp
├── start            # Start MCP server on stdio
├── stream           # Stream MCP server over HTTP
├── tools            # Export available MCP tools as JSON
├── claude
│   ├── enable       # Add server to Claude Desktop config
│   ├── disable      # Remove server from Claude Desktop config
│   └── list         # List Claude Desktop MCP servers
├── vscode
│   ├── enable       # Add server to VSCode config
│   ├── disable      # Remove server from VSCode config
│   └── list         # List VSCode MCP servers
└── cursor
    ├── enable       # Add server to Cursor config
    ├── disable      # Remove server from Cursor config
    └── list         # List Cursor MCP servers
```

## Configuration

Control which commands and flags are exposed as MCP tools using selectors. By default, all commands and flags are exposed (except hidden/deprecated).

```go
config := &ophis.Config{
    Selectors: []ophis.Selector{
        {
            CmdSelector: ophis.AllowCmdsContaining("get", "list"),
            LocalFlagSelector: ophis.ExcludeFlags("token", "secret"),
            InheritedFlagSelector: ophis.NoFlags,  // Exclude persistent flags

            // Middleware wraps command execution
            Middleware: func(ctx context.Context, req *mcp.CallToolRequest, in ophis.ToolInput, next func(context.Context, *mcp.CallToolRequest, ophis.ToolInput) (*mcp.CallToolResult, ophis.ToolOutput, error)) (*mcp.CallToolResult, ophis.ToolOutput, error) {
                ctx, cancel := context.WithTimeout(ctx, time.Minute)
                defer cancel()
                return next(ctx, req, in)
            },
        },
    },
}

rootCmd.AddCommand(ophis.Command(config))
```

See [docs/config.md](docs/config.md) for detailed configuration options.

## How It Works

Ophis bridges Cobra commands and the Model Context Protocol:

1. **Command Discovery**: Recursively walks your Cobra command tree
2. **Schema Generation**: Creates JSON schemas from command flags and arguments ([docs/schema.md](docs/schema.md))
3. **Tool Execution**: Spawns your CLI as a subprocess and captures output ([docs/execution.md](docs/execution.md))

## Contributing

Contributions welcome! See [CONTRIBUTING.md](CONTRIBUTING.md).
