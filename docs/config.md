# Configuration

## Selectors

Selectors control which commands and flags become MCP tools. Ophis evaluates selectors in order and uses the **first matching selector** for each command. Commands that don't match any selector are not exposed.

### Flag Types

Cobra commands have two types of flags:

- **Local Flags**: Defined on a specific command (e.g., `cmd.Flags().String("output", ...)`)
- **Inherited Flags**: Persistent flags defined on parent commands that are available to all subcommands (e.g., `cmd.PersistentFlags().String("config", ...)`)

Ophis allows you to control these separately:
- `LocalFlagSelector`: Controls which local flags are included
- `InheritedFlagSelector`: Controls which persistent flags are included

This separation lets you apply different policies to command-specific flags vs. global configuration flags.

### Default Behavior

- If `Config.Selectors` is nil or empty, all commands and flags are exposed
- If `CmdSelector` is nil, the selector matches all commands
- If `LocalFlagSelector` or `InheritedFlagSelector` is nil, all flags are included

### Automatic Filtering

These are always filtered out regardless of selector configuration:
- Hidden and deprecated commands
- Non-runnable commands (no Run/RunE function)
- Built-in commands (mcp, help, completion)
- Hidden and deprecated flags

## Basic Examples

### Expose specific commands

```go
config := &ophis.Config{
    Selectors: []ophis.Selector{
        {
            CmdSelector: ophis.AllowCmds("kubectl get", "kubectl describe"),
        },
    },
}
```

### Exclude sensitive flags

```go
config := &ophis.Config{
    Selectors: []ophis.Selector{
        {
            LocalFlagSelector: ophis.ExcludeFlags("token", "password", "secret"),
            InheritedFlagSelector: ophis.NoFlags, // Exclude all persistent flags
        },
    },
}
```

### Different rules for different commands

```go
config := &ophis.Config{
    Selectors: []ophis.Selector{
        {
            // Read operations: include all flags, add timeout
            CmdSelector: ophis.AllowCmdsContaining("get", "list", "describe"),
            PreRun: func(ctx context.Context, req *mcp.CallToolRequest, in bridge.ToolInput) (context.Context, *mcp.CallToolRequest, bridge.ToolInput) {
                ctx, _ = context.WithTimeout(ctx, time.Minute)
                return ctx, req, in
            },
        },
        {
            // Write operations: exclude dangerous flags
            CmdSelector: ophis.AllowCmdsContaining("create", "delete", "apply"),
            LocalFlagSelector: ophis.ExcludeFlags("force", "all"),
        },
        {
            // Catch-all: exclude common sensitive flags
            LocalFlagSelector: ophis.ExcludeFlags("token", "kubeconfig"),
        },
    },
}
```

## Built-in Selector Functions

### Command Selectors

- `AllowCmds(cmds ...string)` - Exact command path matches
- `ExcludeCmds(cmds ...string)` - Exclude exact command paths
- `AllowCmdsContaining(substrings ...string)` - Command path contains any substring
- `ExcludeCmdsContaining(substrings ...string)` - Command path excludes all substrings

### Flag Selectors

- `AllowFlags(names ...string)` - Include only these flags
- `ExcludeFlags(names ...string)` - Exclude these flags
- `NoFlags` - Exclude all flags

## Custom Selectors

Use functions for complex logic:

```go
config := &ophis.Config{
    Selectors: []ophis.Selector{
        {
            CmdSelector: func(cmd *cobra.Command) bool {
                // Only expose commands annotated with "mcp"
                return cmd.Annotations["mcp"] == "true"
            },
            LocalFlagSelector: func(flag *pflag.Flag) bool {
                // Include flags based on custom logic
                if flag.Annotations["mcp"] == "true" {
                    return true
                }
                // Exclude flags with "internal" annotation
                return flag.Annotations["internal"] != "true"
            },
        },
    },
}
```

## Middleware Hooks

Add behavior before and after tool execution:

```go
config := &ophis.Config{
    Selectors: []ophis.Selector{
        {
            PreRun: func(ctx context.Context, req *mcp.CallToolRequest, in bridge.ToolInput) (context.Context, *mcp.CallToolRequest, bridge.ToolInput) {
                // Add timeout
                ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
                defer cancel()
                
                // Log tool invocation
                log.Printf("Tool %s called with args: %v", req.Params.Name, in.Args)
                
                // Validate or modify input
                if len(in.Args) > 10 {
                    in.Args = in.Args[:10] // Limit arguments
                }
                
                return ctx, req, in
            },
            PostRun: func(ctx context.Context, req *mcp.CallToolRequest, in bridge.ToolInput, res *mcp.CallToolResult, out bridge.ToolOutput, err error) (*mcp.CallToolResult, bridge.ToolOutput, error) {
                // Log execution result
                log.Printf("Tool %s exited with code %d", req.Params.Name, out.ExitCode)
                
                // Filter or sanitize output
                if strings.Contains(out.StdOut, "SENSITIVE") {
                    out.StdOut = "[REDACTED]"
                }
                
                // Handle errors
                if err != nil {
                    log.Printf("Tool execution failed: %v", err)
                }
                
                return res, out, err
            },
        },
    },
}
```

## Logging

Configure structured logging to stderr:

```go
config := &ophis.Config{
    SloggerOptions: &slog.HandlerOptions{
        Level: slog.LevelDebug,
        AddSource: true,
    },
}
```

Or set via command line:

```bash
./my-cli mcp start --log-level debug
./my-cli mcp claude enable --log-level info
```

## Common Patterns

### Expose only read operations

```go
config := &ophis.Config{
    Selectors: []ophis.Selector{
        {
            CmdSelector: ophis.AllowCmdsContaining("get", "list", "describe", "show"),
            InheritedFlagSelector: ophis.NoFlags,  // Exclude global config flags
        },
    },
}
```

### Different timeout policies per command type

```go
config := &ophis.Config{
    Selectors: []ophis.Selector{
        {
            // Quick operations: 30s timeout
            CmdSelector: ophis.AllowCmdsContaining("get", "describe"),
            PreRun: func(ctx context.Context, req *mcp.CallToolRequest, in bridge.ToolInput) (context.Context, *mcp.CallToolRequest, bridge.ToolInput) {
                ctx, _ = context.WithTimeout(ctx, 30*time.Second)
                return ctx, req, in
            },
        },
        {
            // Long operations: 5m timeout
            CmdSelector: ophis.AllowCmdsContaining("apply", "create", "install"),
            PreRun: func(ctx context.Context, req *mcp.CallToolRequest, in bridge.ToolInput) (context.Context, *mcp.CallToolRequest, bridge.ToolInput) {
                ctx, _ = context.WithTimeout(ctx, 5*time.Minute)
                return ctx, req, in
            },
        },
    },
}
```

### Exclude all inherited flags globally

```go
config := &ophis.Config{
    Selectors: []ophis.Selector{
        {
            // Matches all commands, excludes all persistent flags
            InheritedFlagSelector: ophis.NoFlags,
            LocalFlagSelector: ophis.ExcludeFlags("token", "password"),
        },
    },
}
```

### Annotation-based exposure

```go
// In your CLI code, annotate commands you want to expose:
getCmd.Annotations = map[string]string{"mcp": "true"}

// In your config:
config := &ophis.Config{
    Selectors: []ophis.Selector{
        {
            CmdSelector: func(cmd *cobra.Command) bool {
                return cmd.Annotations["mcp"] == "true"
            },
        },
    },
}
```

### Sanitize output for security

```go
config := &ophis.Config{
    Selectors: []ophis.Selector{
        {
            PostRun: func(ctx context.Context, req *mcp.CallToolRequest, in bridge.ToolInput, res *mcp.CallToolResult, out bridge.ToolOutput, err error) (*mcp.CallToolResult, bridge.ToolOutput, error) {
                // Remove sensitive patterns from output
                patterns := []string{
                    "token:",
                    "password:",
                    "secret:",
                    "api-key:",
                }
                for _, pattern := range patterns {
                    out.StdOut = regexp.MustCompile(pattern+`[^\s]+`).ReplaceAllString(out.StdOut, pattern+"[REDACTED]")
                    out.StdErr = regexp.MustCompile(pattern+`[^\s]+`).ReplaceAllString(out.StdErr, pattern+"[REDACTED]")
                }
                return res, out, err
            },
        },
    },
}
```
