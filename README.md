![Project Logo](./logo.png)

**Transform any Cobra CLI into an MCP server**

Ophis automatically converts your existing Cobra commands into MCP tools, and provides CLI commands for integration with Claude and VSCode.

## Import

```bash
go get github.com/njayp/ophis
```

## Quick Start

### Add MCP server commands to your command tree.

MCP commands can be added anywhere in a command tree. Below is an example of a `main()` that adds MCP commands to a root command. Alternatively, this logic can be placed in your `createMyRootCommand()`.

```go
package main

import (
    "os"
    "github.com/njayp/ophis"
)

func main() {
    rootCmd := createMyRootCommand()
    
    // Add MCP server commands
    rootCmd.AddCommand(ophis.Command(nil))
    
    if err := rootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}
```

### Enable in Claude Desktop or VSCode

```bash
./my-cli mcp claude enable
# Restart Claude Desktop
```

```bash
./my-cli mcp vscode enable
# Ensure Copilot is in Agent Mode
```

Your CLI commands are now available as mcp server tools!

## Configuration

The `ophis.Command()` function accepts an optional `*ophis.Config` parameter to customize the MCP server behavior:

```go
config := &ophis.Config{
    // Customize command filtering
    Filters: []ophis.Filter{
        // Command filtering
        ophis.ExcludeFilter([]string{"dangerous"}),
    },
    
    // Configure logging (logs to stderr)
    SloggerOptions: &slog.HandlerOptions{
        Level: slog.LevelDebug,
    },
}
```

### Command Filtering

Control which commands are exposed as MCP tools:

```go
// only exposes listed commands
ophis.AllowFilter([]string{"get", "list", "helm repo list"})
```

```go
// prevents listed commands from being exposed
ophis.ExcludeFilter([]string{"delete", "destroy", "helm repo remove"})
```

```go
// Custom filter function
func(cmd *cobra.Command) bool {
    // Exclude admin commands
    return !strings.HasPrefix(cmd.Name(), "admin-")
}
```

When `Config.Filters` is `nil`, ophis uses these default filters:
- Excludes commands without a `Run` or `PreRun` function
- Excludes hidden commands
- Excludes `mcp`, `help`, and `completion` commands

## Ophis Commands

`ophis.Command` returns the following commands:

```
mcp
├── start            # Start MCP server on stdio
├── tools            # Export available MCP tools as JSON
├── claude
│   ├── enable       # Enable Helm MCP in Claude Desktop
│   ├── disable      # Disable Helm MCP in Claude Desktop
│   └── list         # List MCP configurations in Claude Desktop
└── vscode
    ├── enable       # Enable Helm MCP in VS Code
    ├── disable      # Disable Helm MCP in VS Code
    └── list         # List MCP configurations in VS Code
```

## Examples

Run `make build` to build all examples to `ophis/bin`.

- [kubectl](./examples/kubectl/main.go)
- [make](./examples/make/)

### External Examples

- [helm](https://github.com/njayp/helm)

## Contributing

Contributions welcome! See [CONTRIBUTING.md](CONTRIBUTING.md)