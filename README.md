# Ophis

**Transform any Cobra CLI into an MCP server**

Ophis automatically converts your existing Cobra commands into MCP tools that Claude can use.

## Installation

```go
go get github.com/njayp/ophis
```

## Quick Start

### Add MCP server commands to your root command.

Below is an example of a `main()` that adds MCP commands to a root command. Alternatively, this logic can be placed in your `createMyRootCommand()`.

```go
package main

import (
    "os"
    "github.com/njayp/ophis/bridge"
    "github.com/njayp/ophis/mcp"
)

func main() {
    rootCmd := createMyRootCommand()
    
    // Add MCP server commands
    rootCmd.AddCommand(mcp.Command(&bridge.Config{
        AppName: "my-app-name",
        RootCmd: command,
    }))
    
    if err := rootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}
```

### Enable in Claude Desktop

```bash
./my-cli mcp claude enable
```

**Restart Claude Desktop**

Your CLI commands are now available as tools in Claude!

## Features

- Automatic command-to-tool conversion
- Full flag and argument support
- Preserves command descriptions and help text
- Zero changes to existing commands
- Selective command filtering
- Spawns a new process for each tool call

## Examples

- [helm](https://github.com/njayp/helm)
- [kubectl](https://github.com/njayp/kubectl)

## Contributing

Contributions welcome! See [CONTRIBUTING.md](CONTRIBUTING.md)