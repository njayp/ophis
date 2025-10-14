# Tool Execution

When an AI assistant calls an MCP tool, Ophis executes your CLI command and returns the results.

## Execution Flow

1. **PreRun Middleware** (optional)
   - Runs before command execution
   - Can modify context, request, or input
   - Can cancel execution by returning cancelled context
   - Common uses: timeouts, validation, logging, auth checks

2. **Command Execution**
   - Spawns CLI as subprocess with constructed arguments
   - Captures stdout, stderr, and exit code
   - Runs with same permissions as MCP server process

3. **PostRun Middleware** (optional)
   - Runs after command execution
   - Can modify result, output, or error
   - Common uses: output filtering, error handling, metrics

## Command Construction

Tool calls are converted to CLI invocations:

### Input

```json
{
  "name": "kubectl_get_pods",
  "arguments": {
    "flags": {
      "namespace": "production",
      "output": "json",
      "verbose": true
    },
    "args": ["web-server"]
  }
}
```

### Constructed Command

```bash
/path/to/kubectl get pods --namespace production --output json --verbose web-server
```

### Flag Conversion Rules

- **Boolean flags**: 
  - `true` → `--flag-name`
  - `false` → omitted
  
- **String/numeric flags**: `--flag-name value`

- **Array flags**: Repeated for each element
  - `{"labels": ["a", "b"]}` → `--labels a --labels b`

- **Empty/null values**: Omitted

## Output Format

All executions return:

```json
{
  "stdout": "command output...",
  "stderr": "error messages...",
  "exitCode": 0
}
```

- **stdout**: Everything written to standard output
- **stderr**: Everything written to standard error
- **exitCode**: Process exit code (0 = success, non-zero = error)

## Error Handling

Ophis distinguishes between:

1. **Execution errors**: Command failed to start (returns error, no output)
2. **Command errors**: Command ran but returned non-zero exit code (returns output with exitCode)

Example of command error:

```json
{
  "stdout": "",
  "stderr": "Error: namespace 'invalid' not found\n",
  "exitCode": 1
}
```

## Context and Cancellation

The execution context can be cancelled:

- By PreRun middleware returning a cancelled context
- By the MCP client cancelling the request
- By the parent context timing out

Cancelled executions kill the subprocess and return an error.

## Subprocess Details

Commands are executed with:

- **Working directory**: Same as MCP server process
- **Environment**: Inherited from MCP server process
- **Permissions**: Same as MCP server process
- **Streams**: stdout and stderr captured, stdin closed

## Middleware Examples

### Timeout

```go
PreRun: func(ctx context.Context, req *mcp.CallToolRequest, in bridge.ToolInput) (context.Context, *mcp.CallToolRequest, bridge.ToolInput) {
    ctx, _ = context.WithTimeout(ctx, 30*time.Second)
    return ctx, req, in
}
```

### Input Validation

```go
PreRun: func(ctx context.Context, req *mcp.CallToolRequest, in bridge.ToolInput) (context.Context, *mcp.CallToolRequest, bridge.ToolInput) {
    // Limit number of arguments
    if len(in.Args) > 10 {
        ctx, cancel := context.WithCancel(ctx)
        cancel() // Prevent execution
        return ctx, req, in
    }
    return ctx, req, in
}
```

### Output Filtering

```go
PostRun: func(ctx context.Context, req *mcp.CallToolRequest, in bridge.ToolInput, res *mcp.CallToolResult, out bridge.ToolOutput, err error) (*mcp.CallToolResult, bridge.ToolOutput, error) {
    // Redact sensitive information
    out.StdOut = strings.ReplaceAll(out.StdOut, secretToken, "[REDACTED]")
    out.StdErr = strings.ReplaceAll(out.StdErr, secretToken, "[REDACTED]")
    return res, out, err
}
```

### Logging

```go
PreRun: func(ctx context.Context, req *mcp.CallToolRequest, in bridge.ToolInput) (context.Context, *mcp.CallToolRequest, bridge.ToolInput) {
    log.Printf("[MCP] Executing: %s %v", req.Params.Name, in.Args)
    return ctx, req, in
}

PostRun: func(ctx context.Context, req *mcp.CallToolRequest, in bridge.ToolInput, res *mcp.CallToolResult, out bridge.ToolOutput, err error) (*mcp.CallToolResult, bridge.ToolOutput, error) {
    log.Printf("[MCP] Completed: %s (exit=%d)", req.Params.Name, out.ExitCode)
    return res, out, err
}
```

### Rate Limiting

```go
var limiter = rate.NewLimiter(rate.Every(time.Second), 10)

PreRun: func(ctx context.Context, req *mcp.CallToolRequest, in bridge.ToolInput) (context.Context, *mcp.CallToolRequest, bridge.ToolInput) {
    if err := limiter.Wait(ctx); err != nil {
        ctx, cancel := context.WithCancel(ctx)
        cancel()
        return ctx, req, in
    }
    return ctx, req, in
}
```
