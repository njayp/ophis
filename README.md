![Project Logo](./logo.png)

**Transform any Cobra CLI into an MCP server**

Ophis automatically converts your existing Cobra commands into MCP tools that Claude can use.

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
import (
    "log/slog"
    "github.com/njayp/ophis"
    "github.com/njayp/ophis/tools"
)

config := &ophis.Config{
    // Customize command filtering and output handling
    GeneratorOptions: []tools.GeneratorOption{
        // Command filtering
        tools.AddFilter(tools.Exclude([]string{"dangerous"})),
        
        // Custom output handler
        tools.WithHandler(myCustomHandler),
    },
    
    // Configure logging (logs to stderr)
    SloggerOptions: &slog.HandlerOptions{
        Level: slog.LevelDebug,
    },
}

rootCmd.AddCommand(ophis.Command(config))
```

### Default Behavior

When called with `nil` config, the MCP server:
- Excludes commands without a Run or PreRun function
- Excludes hidden, "mcp", "help", and "completion" commands
- Returns command output as plain text
- Logs at info level

### Command Filtering

Control which commands are exposed as MCP tools:

```go
// Only expose specific commands
tools.WithFilters(tools.Allow([]string{"get", "list", "describe"}))
```

```go
// Or exclude specific commands (in addition to defaults)
tools.AddFilter(tools.Exclude([]string{"delete", "destroy"}))
```

```go
// Custom filter function
tools.AddFilter(func(cmd *cobra.Command) bool {
    // Exclude admin commands
    return !strings.HasPrefix(cmd.Name(), "admin-")
})
```

### Custom Output Handler

The output handler will be applied to all tools. See proposal [#9](https://github.com/njayp/ophis/issues/9).

```go
// Return the data as an image instead of as text
tools.WithHandler(func(ctx context.Context, request mcp.CallToolRequest, data []byte, err error) *mcp.CallToolResult {
    return mcp.NewToolResultImage(data)
})
```

```go
// Or add middleware
tools.WithHandler(func(ctx context.Context, request mcp.CallToolRequest, data []byte, err error) *mcp.CallToolResult {
    // Your middleware here
    return tools.DefaultHandler(ctx, request, data, err)
})
```

## Ophis Commands

`ophis.Command` returns the following tree of commands:

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

- [helm](https://github.com/njayp/helm)
- [kubectl](https://github.com/njayp/kubectl)

## Contributing

Contributions welcome! See [CONTRIBUTING.md](CONTRIBUTING.md)