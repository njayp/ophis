![Project Logo](./logo.png)

**Transform any Cobra CLI into an MCP server**

Ophis automatically converts your existing Cobra commands into MCP tools that Claude can use.

## Import

```bash
go get github.com/njayp/ophis
```

## Quick Start

### Add MCP server commands to your root command.

Below is an example of a `main()` that adds MCP commands to a root command. Alternatively, this logic can be placed in your `createMyRootCommand()`.

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
```

```bash
./my-cli mcp vscode enable
```

**Restart Claude Desktop or VSCode**

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
    Generator: tools.NewGenerator(
        // Include only specific commands
        tools.WithFilters(tools.Allow([]string{"get", "list"})),
        
        // Or exclude specific commands
        tools.AddFilter(tools.Exclude([]string{"dangerous"})),
        
        // Custom output handler
        tools.WithHandler(myCustomHandler),
    ),
    
    // Configure logging (logs to stderr)
    SloggerOptions: &slog.HandlerOptions{
        Level: slog.LevelDebug,
    },
}

rootCmd.AddCommand(ophis.Command(config))
```

### Default Behavior

When called with `nil` config, the MCP server:
- Excludes hidden, "mcp", "help", and "completion" commands
- Returns command output as plain text
- Logs at info level

### Command Filtering

Control which commands are exposed as MCP tools:

```go
// Only expose specific commands
tools.WithFilters(tools.Allow([]string{"get", "list", "describe"}))

// Exclude specific commands (in addition to defaults)
tools.AddFilter(tools.Exclude([]string{"delete", "destroy"}))

// Custom filter function
tools.AddFilter(func(cmd *cobra.Command) bool {
    // Exclude admin commands
    return !strings.HasPrefix(cmd.Name(), "admin-")
})
```

### Custom Output Handler

Return the data as an image instead of as text.

```go
tools.WithHandler(func(request mcp.CallToolRequest, data []byte) *mcp.CallToolResult {
    return mcp.NewToolResultImage(data)
})
```

## Examples

- [helm](https://github.com/njayp/helm)
- [kubectl](https://github.com/njayp/kubectl)

## Contributing

Contributions welcome! See [CONTRIBUTING.md](CONTRIBUTING.md)