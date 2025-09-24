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

Ophis uses a powerful selector system to give you fine-grain control over which commands and flags are exposed as MCP tools. 

**Select Commands:**
```go
type Selector struct {
	// Selects *cobra.Commands to be made into MCP tools
	CmdSelector CmdSelector
	// Selects flags for selected *cobra.Commands
	FlagSelector FlagSelector
}
```

The first selector that matches a command will convert that command into a MCP tool. If no selector matches a command, it will not be made into a tool.

#### Default

If `Config.Selectors` is nil, all valid commands will be converted into MCP tools.

#### Basic Example

```go
// Expose only specific read operations
config := &ophis.Config{
    Selectors: []ophis.Selector{
        {
            CmdSelector: ophis.AllowCmd("get", "helm repo list"),
            // FlagSelector defaults to all non-hidden, non-deprecated flags
        },
    },
}
```

#### Advanced: Different Flag Rules for Different Commands

The real power comes from combining multiple selectors with different flag rules. See the helm example [config](./examples/helm/config.go).

#### Custom Selector Functions

For complex logic, use custom functions:

```go
ophis.Selector{
    // Match commands based on custom logic (basic exclusions still apply)
    CmdSelector: func(cmd *cobra.Command) bool {
        // Only expose commands that have been annotated as "safe"
        return cmd.Annotations["mcp-safe"] == "true"
    },
    // Include only flags that don't modify state (hidden/deprecated still excluded)
    FlagSelector: func(flag *pflag.Flag) bool {
        return flag.Annotations["mcp-safe"] == "true"
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