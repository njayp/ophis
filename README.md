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
    // Customize command selection
    Selectors: []ophis.Selector{
        {
            CmdSelect: ophis.ExcludeCmd("dangerous"),
        },
    },

    // Send metrics or set timeouts
    PreRun: func(ctx context.Context, req *mcp.CallToolRequest, in bridge.CmdToolInput) (context.Context, *mcp.CallToolRequest, bridge.CmdToolInput) {
        // your middleware here
        ctx, _ = context.WithTimeout(ctx, time.Minute)
        return ctx, req, in
    },
    
    // Configure logging (logs to stderr)
    SloggerOptions: &slog.HandlerOptions{
        Level: slog.LevelDebug,
    },
}
```

### Command and Flag Selection

Ophis uses a powerful selector system to give you fine-grained control over which commands and flags are exposed as MCP tools. Each selector is a rule that defines:
- Which commands to match (`CmdSelect`)
- Which flags to include for those matched commands (`FlagSelect`)

**How it works:** Selectors are evaluated in order. The first selector whose `CmdSelect` matches a command wins, and its `FlagSelect` determines which flags are included for that command.

#### Basic Examples

```go
// Expose only safe read operations
ophis.Selector{
    CmdSelect: ophis.AllowCmd("get", "list", "describe"),
    // FlagSelect defaults to including all non-hidden flags
}
```

```go
// Exclude dangerous operations
ophis.Selector{
    CmdSelect: ophis.ExcludeCmd("delete", "destroy", "remove"),
    // Commands that don't match are not exposed at all
}
```

#### Advanced: Different Flag Rules for Different Commands

The real power comes from combining multiple selectors with different flag rules:

```go
config := &ophis.Config{
    Selectors: []ophis.Selector{
        {
            // For 'get' commands: include only output formatting flags
            CmdSelect: ophis.AllowCmd("get"),
            FlagSelect: ophis.AllowFlag("output", "namespace", "selector"),
        },
        {
            // For 'delete' commands: exclude dangerous flags
            CmdSelect: ophis.AllowCmd("delete"),
            FlagSelect: ophis.ExcludeFlag("all", "force", "grace-period"),
        },
        {
            // For 'create/apply' commands: include most flags but exclude auth-related ones
            CmdSelect: ophis.AllowCmd("create", "apply"),
            FlagSelect: ophis.ExcludeFlag("token", "kubeconfig", "context"),
        },
        {
            // Default for all other commands: basic safety exclusions
            CmdSelect: func(cmd *cobra.Command) bool { 
                return !strings.Contains(cmd.CommandPath(), "admin")
            },
            FlagSelect: ophis.ExcludeFlag("insecure", "tls-skip"),
        },
    },
}
```

#### Custom Selector Functions

For complex logic, use custom functions:

```go
ophis.Selector{
    // Match commands based on custom logic
    CmdSelect: func(cmd *cobra.Command) bool {
        // Only expose commands that have been annotated as "safe"
        return cmd.Annotations["mcp-safe"] == "true"
    },
    // Include only flags that don't modify state
    FlagSelect: func(flag *pflag.Flag) bool {
        return !strings.Contains(flag.Usage, "delete") && 
               !strings.Contains(flag.Usage, "remove") &&
               flag.Name != "force"
    },
}
```

#### Selector Evaluation Order

1. Selectors are evaluated in the order they appear in the slice
2. The first selector whose `CmdSelect` returns `true` for a command wins
3. That selector's `FlagSelect` (if provided) determines which flags are included
4. If no selectors match a command, it's excluded from MCP tools

The following commands are always excluded:
- Commands without a `Run` or `PreRun` function
- Hidden commands
- Deprecated commands
- `mcp`, `help`, and `completion` commands

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