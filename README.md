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
    // Customize tool creation and execution
    Selectors: []ophis.Selector{
        {
            CmdSelector: ophis.ExcludeCmd("dangerous"),
            PreRun: func(ctx context.Context, req *mcp.CallToolRequest, in bridge.CmdToolInput) (context.Context, *mcp.CallToolRequest, bridge.CmdToolInput) {
                // Add timeout
                ctx, _ = context.WithTimeout(ctx, time.Minute)
                return ctx, req, in
            },
        },
    },
    
    // Configure logging (logs to stderr)
    SloggerOptions: &slog.HandlerOptions{
        Level: slog.LevelDebug,
    },
}
```

### Selectors

Selectors control which commands become MCP tools and which flags they include.

#### Basic Examples

```go
// Default: expose all commands with all their flags
config := &ophis.Config{}
```

```go
// Basic selection
config := &ophis.Config{
    Selectors: []ophis.Selector{
        {
            // Only these commands
            CmdSelector: ophis.AllowCmd("get", "helm repo list"),
            // Without this flag
            FlagSelector: ophis.ExcludeFlag("kubeconfig"),
        },
    },
}
```

```go
config := &ophis.Config{
    Selectors: []ophis.Selector{
        {
            // Only these commands, with all flags
            CmdSelector: ophis.AllowCmd("get", "helm repo list"),
        },
    },
}
```

```go
config := &ophis.Config{
    Selectors: []ophis.Selector{
        {
            // All commands, without this flag
            FlagSelector: ophis.ExcludeFlag("kubeconfig"),
        },
    },
}
```

#### How Selectors Select Commands

1. **Safety first**: Hidden, deprecated, and non-runnable commands/flags are always excluded
2. **First match wins**: Selectors are evaluated in order; the first matching `CmdSelector` determines which `FlagSelector` is used
3. **No match = no tool**: Commands that don't match any selector are not exposed

#### Middleware Hooks

Each selector can include middleware hooks that run before and after tool execution. Different selectors can specify different PreRun and PostRun functions for different commands.

```go
config := &ophis.Config{
    Selectors: []ophis.Selector{
        {
            // PreRun executes before each tool call
            // Return a cancelled context to prevent execution
            PreRun: func(ctx context.Context, req *mcp.CallToolRequest, in bridge.CmdToolInput) (context.Context, *mcp.CallToolRequest, bridge.CmdToolInput) {
                // Add timeout
                ctx, _ = context.WithTimeout(ctx, time.Minute)
                return ctx, req, in
            },
            
            // PostRun executes after each tool call
            PostRun: func(ctx context.Context, req *mcp.CallToolRequest, in bridge.CmdToolInput, res *mcp.CallToolResult, out bridge.CmdToolOutput, err error) (*mcp.CallToolResult, bridge.CmdToolOutput, error) {
                // Your middleware here
                return res, out, err
            },
        },
    },
}
```

Common use cases for middleware:
- **PreRun**: Add timeouts, rate limiting, authentication checks, request logging
- **PostRun**: Error handling, response filtering, metrics collection, output sanitization

#### Multiple Selectors

Different commands can have different flag rules and execution hooks:

```go
config := &ophis.Config{
    Selectors: []ophis.Selector{
        {
            // Read operations: only output flags
            CmdSelector: ophis.AllowCmd("get", "list", "logs"),
            FlagSelector: ophis.AllowFlag("output", "format"),
            PreRun: timeoutFn(),
            PostRun: limitOutputFn(),
        },
        {
            // Write operations: exclude dangerous flags
            CmdSelector: ophis.AllowCmd("create", "apply"),
            FlagSelector: ophis.ExcludeFlag("force", "token", "insecure"),
            PreRun: timeoutFn(),
        },
        {
            // Everything else: with common flag exclusions
            FlagSelector: ophis.ExcludeFlag("token", "insecure"),
            PreRun: timeoutFn(),
        },
    },
}
```

#### Custom Selector Functions

For complex logic, use custom functions:

```go
ophis.Selector{
    // Match commands based on custom logic
    CmdSelector: func(cmd *cobra.Command) bool {
        // Only expose commands that have been annotated as "mcp"
        return cmd.Annotations["mcp"] == "true"
    },
    FlagSelector: func(flag *pflag.Flag) bool {
        return flag.Annotations["mcp"] == "true"
    },
}
```


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

- [kubectl](./examples/kubectl/)
- [helm](./examples/helm/)
- [make](./examples/make/)

### External Examples

- [helm](https://github.com/njayp/helm)

## Contributing

Contributions welcome! See [CONTRIBUTING.md](CONTRIBUTING.md)