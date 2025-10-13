![Project Logo](./logo.png)

**Transform any Cobra CLI into an MCP server**

Ophis automatically converts your existing Cobra commands into MCP tools, and provides CLI commands for integration with Claude and VSCode.

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

## Quick Start

### Import

```bash
go get github.com/njayp/ophis
```

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

## Config

`Selectors` select commands and flags to be made into MCP tools, and provide hooks for PreRun and PostRun middleware for those tools.

### How Selectors Work

Selectors control which commands and flags become MCP tools. Each selector contains a `CmdSelector` that chooses which commands to expose and `LocalFlagSelector`, `InheritedFlagSelector` that determine which flags to include. When converting commands, ophis evaluates selectors in order and uses the **first matching selector** to create each command's MCP tool. Commands that don't match any selector are not exposed.

If `CmdSelector` is nil, all commands are allowed. If a flag selector is nil, all flags are allowed. This makes it easy to create catch-all selectors or to expose everything by default.

Hidden, deprecated, and non-runnable commands are automatically filtered out, as are hidden and deprecated flags.

Each selector can also include PreRun and PostRun middleware hooks that execute before and after tool invocation. This lets you apply different behavior to different commands—for example, adding timeouts to read operations, excluding sensitive flags from write operations, or sanitizing outputs for public-facing tools.

### Multiple Selectors Example

Different commands can have different flag rules and execution hooks:

```go
config := &ophis.Config{
    Selectors: []ophis.Selector{
        {
            // Select all get commands
            CmdSelector: ophis.AllowCmdsContaining("get"),

            // LocalFlagSelector is nil, selecting all local flags

            // Exclude all inherited flags
            InheritedFlagSelector: ophis.NoFlags,

            // Add middleware for these commands
            PreRun: func(ctx context.Context, req *mcp.CallToolRequest, in bridge.CmdToolInput) (context.Context, *mcp.CallToolRequest, bridge.CmdToolInput) {
                // Add timeout
                ctx, _ = context.WithTimeout(ctx, time.Minute)
                return ctx, req, in
            },
        },
        {
            // Select exactly these write commands
            CmdSelector: ophis.AllowCmds("mycli create", "mycli apply"),

            // Exclude dangerous flags
            LocalFlagSelector: ophis.ExcludeFlags("force", "token", "insecure"),

            // Exclude all inherited flags
            InheritedFlagSelector: ophis.NoFlags,
        },
        {
            // CmdSelector is nil, selecting all remaining commands

            // Exclude common, dangerous local flags
            LocalFlagSelector: ophis.ExcludeFlags("token", "insecure"),

            // Exclude all inherited flags
            InheritedFlagSelector: ophis.NoFlags,
        },
    },
}
```

### Custom Selector Functions

For complex logic, use custom functions:

```go
ophis.Selector{
    // Match commands based on custom logic
    CmdSelector: func(cmd *cobra.Command) bool {
        // Only expose commands that have been annotated as "mcp"
        return cmd.Annotations["mcp"] == "true"
    },

    LocalFlagSelector: func(flag *pflag.Flag) bool {
        return flag.Annotations["mcp"] == "true"
    },

    InheritedFlagSelector: func(flag *pflag.Flag) bool {
        return flag.Annotations["mcp"] == "true"
    },
}
```

## Under the Hood

Ophis bridges Cobra commands and the Model Context Protocol through three main components:

### Command Discovery and Filtering

Ophis recursively walks your Cobra command tree at runtime, creating MCP tools from selected commands.

### Schema Generation

For each selected command, ophis automatically generates JSON schemas for both input and output:

**Input Schema** includes:

- `flags`: An object with properties for each included flag, typed according to Cobra's flag types (string, bool, int, stringSlice, etc.)
- `args`: An array of strings for positional arguments, enhanced with usage patterns from the command's `Use` field

Flag schemas include rich metadata like descriptions, default values, required status, and format constraints. This allows AI assistants to understand exactly how to invoke each tool.

**Output Schema** includes:

- `stdout`: Captured standard output
- `stderr`: Captured standard error
- `exitCode`: The command's exit code

### Tool Execution

When an AI assistant calls an MCP tool, ophis runs optional `PreRun` middleware, spawns your CLI as a subprocess with the requested flags and arguments, captures stdout/stderr/exitCode, then runs optional `PostRun` middleware. Your CLI is invoked exactly as a user would from the command line.

## Examples

Run `make build` to build all examples to `ophis/bin`.

- [kubectl](./examples/kubectl/)
- [helm](./examples/helm/)
- [argocd](./examples/argocd/)
- [make](./examples/make/)

## Contributing

Contributions welcome! See [CONTRIBUTING.md](CONTRIBUTING.md)
