# Configuration

## Selectors

Selectors control which commands and flags become MCP tools. Ophis evaluates selectors in order and uses the **first matching selector** for each command. If no selectors match a command, the command is not exposed as a MCP tool.

### Default Behavior

- If `Config.Selectors` is nil/empty, all commands and flags are exposed
- If `CmdSelector` is nil, the selector matches all commands
- If `LocalFlagSelector` or `InheritedFlagSelector` is nil, all flags are included

### Automatic Filtering

Always filtered regardless of configuration:

- Hidden and deprecated commands/flags
- Non-runnable commands (no Run/RunE)
- Built-in commands (mcp, help, completion)

## Examples

### Expose Specific Commands

```go
config := &ophis.Config{
    Selectors: []ophis.Selector{
        {
            CmdSelector: ophis.AllowCmds("kubectl get", "kubectl describe"),
        },
    },
}
```

### Exclude Sensitive Flags

```go
config := &ophis.Config{
    Selectors: []ophis.Selector{
        {
            LocalFlagSelector: ophis.ExcludeFlags("token", "password"),
            InheritedFlagSelector: ophis.NoFlags,
        },
    },
}
```

### Different Rules per Command Type

```go
config := &ophis.Config{
    Selectors: []ophis.Selector{
        {
            // Read ops: timeout
            CmdSelector: ophis.AllowCmdsContaining("get", "list"),
            PreRun: func(ctx context.Context, req *mcp.CallToolRequest, in bridge.ToolInput) (context.Context, *mcp.CallToolRequest, bridge.ToolInput) {
                ctx, _ = context.WithTimeout(ctx, time.Minute)
                return ctx, req, in
            },
        },
        {
            // Write ops: restrict flags
            CmdSelector: ophis.AllowCmdsContaining("delete", "apply"),
            LocalFlagSelector: ophis.ExcludeFlags("force", "all"),
        },
    },
}
```

## Selector Functions

### Commands

- `AllowCmds(cmds ...string)` - Exact matches
- `ExcludeCmds(cmds ...string)` - Exact exclusions
- `AllowCmdsContaining(substrings ...string)` - Contains any
- `ExcludeCmdsContaining(substrings ...string)` - Excludes all

### Flags

- `AllowFlags(names ...string)` - Include only these
- `ExcludeFlags(names ...string)` - Exclude these
- `NoFlags` - Exclude all

### Custom Selector Functions

```go
config := &ophis.Config{
    Selectors: []ophis.Selector{
        {
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
        },
    },
}
```

## Middleware

Add behavior before/after execution:

```go
PreRun: func(ctx context.Context, req *mcp.CallToolRequest, in bridge.ToolInput) (context.Context, *mcp.CallToolRequest, bridge.ToolInput) {
    // Add timeout
    ctx, _ = context.WithTimeout(ctx, 30*time.Second)

    // Validate input
    if len(in.Args) > 10 {
        in.Args = in.Args[:10]
    }

    return ctx, req, in
}

PostRun: func(ctx context.Context, req *mcp.CallToolRequest, in bridge.ToolInput, res *mcp.CallToolResult, out bridge.ToolOutput, err error) (*mcp.CallToolResult, bridge.ToolOutput, error) {
    // Filter output
    if strings.Contains(out.StdOut, "SECRET") {
        out.StdOut = "[REDACTED]"
    }

    return res, out, err
}
```

## Logging

```go
config := &ophis.Config{
    SloggerOptions: &slog.HandlerOptions{
        Level: slog.LevelDebug,
    },
}
```

Or via CLI:

```bash
./my-cli mcp vscode enable --log-level debug
```
