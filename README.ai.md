![Project Logo](./logo.png)

**Transform any Cobra CLI application into an MCP (Model Context Protocol) server**

Ophis is a Go library that automatically converts Cobra-based command-line applications into MCP servers, allowing AI assistants and other MCP clients to interact with your CLI tools through structured protocols.

## Table of Contents

- [Motivation](#motivation)
- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [How It Works](#how-it-works)
  - [Architecture Overview](#architecture-overview)
  - [Command Discovery](#command-discovery)
  - [Tool Registration](#tool-registration)
  - [Execution Flow](#execution-flow)
- [Claude Desktop Integration](#claude-desktop-integration)
  - [Commands](#commands)
  - [Configuration Options](#configuration-options)
  - [Setup Workflow](#setup-workflow)
- [API Reference](#api-reference)
  - [CommandFactory Interface](#commandfactory-interface)
  - [Config Structure](#config-structure)
  - [Type Mapping](#type-mapping)
- [Project Structure](#project-structure)
- [Examples](#examples)
  - [Basic Make Tool](#basic-make-tool)
  - [Custom CLI Tool](#custom-cli-tool)
- [Advanced Usage](#advanced-usage)
  - [Custom Command Execution](#custom-command-execution)
  - [Error Handling](#error-handling)
  - [Logging Configuration](#logging-configuration)
- [Security Considerations](#security-considerations)
- [Troubleshooting](#troubleshooting)
- [Contributing](#contributing)
- [License](#license)
- [Related Projects](#related-projects)

## Motivation

- Transform your Cobra CLI to an MCP server with **one command**
- Expose CLIs to AI assistants without exposing shell access
- Bridge the gap between command-line tools and AI workflows
- Maintain security by controlling exactly which commands are exposed

## Features

- **Automatic Tool Registration**: Recursively discovers and registers all Cobra commands as MCP tools
- **Flag Support**: Maps Cobra flags to MCP tool parameters with proper type detection
- **Positional Arguments**: Handles command-line arguments through a dedicated parameter
- **Command Hierarchy**: Preserves the structure of nested subcommands
- **Clean State Management**: Each command execution uses a fresh command instance
- **Claude Desktop Integration**: Built-in commands to enable/disable MCP servers in Claude Desktop
- **Type-Safe Parameter Mapping**: Intelligent detection of flag types (string, int, bool, arrays, etc.)
- **Cross-Platform Support**: Works on Linux, macOS, and Windows
- **Concurrent Execution**: Thread-safe command execution with proper isolation

## Installation

```bash
go get github.com/njayp/ophis
```

## Quick Start

See [README](README.md) for basic setup instructions.

## How It Works

### Architecture Overview

Ophis uses a bridge pattern to connect Cobra CLI applications with the MCP protocol:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   MCP Client    │───▶│  Ophis Bridge   │───▶│  Cobra Commands │
│  (Claude, etc.) │    │   (Manager)     │    │   (Your CLI)    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Command Discovery

1. **Command Tree Traversal**: The `tools.FromRootCmd()` function recursively walks through your Cobra command tree
2. **Filtering**: Hidden commands and the `mcp` command itself are excluded from registration
3. **Metadata Extraction**: Command descriptions, flags, and usage information are captured

### Tool Registration

Each executable command becomes an MCP tool with automatically generated metadata:

- **Tool Name**: Derived from command path (e.g., `root_sub_command`)
- **Description**: Uses command's `Long` description, falling back to `Short`
- **Parameters**: Flags mapped to typed parameters, plus a special `args` parameter for positional arguments

### Execution Flow

1. **Fresh Instance Creation**: For each tool call, creates a new command instance via `CommandFactory.New()`
2. **Parameter Application**: Maps MCP parameters back to Cobra flags and arguments
3. **Command Execution**: Runs the command in an isolated context
4. **Output Capture**: Returns stdout/stderr as MCP tool result

Key benefits of this pattern:
- **Isolation**: Each execution gets a fresh command instance
- **State Safety**: Prevents state pollution between tool calls
- **Flexibility**: Allows custom command creation and execution logic

## Claude Desktop Integration

Ophis provides built-in commands for managing Claude Desktop MCP server configuration:

### Commands

- `mcp claude enable` - Add your CLI as an MCP server in Claude Desktop
- `mcp claude disable` - Remove your CLI from Claude Desktop configuration  
- `mcp claude list` - Show all configured MCP servers

### Configuration Options

- `--server-name` - Custom name for the MCP server (default: executable name)
- `--log-level` - Set logging level (debug, info, warn, error)
- `--log-file` - Custom log file path (default: user cache directory)
- `--config-path` - Custom Claude config file path

### Setup Workflow

1. **Build your application** with Ophis integration
2. **Enable MCP server**:
   ```bash
   ./your-cli mcp claude enable --log-level debug
   ```
3. **Restart Claude Desktop** to load the new configuration
4. **Verify setup**:
   ```bash
   ./your-cli mcp claude list
   ```

## API Reference

### CommandFactory Interface

The core interface for integrating Ophis with your Cobra application:

```go
type CommandFactory interface {
    // Tools returns all available MCP tools from your command tree
    Tools() []tools.Tool
    
    // New creates a fresh command instance and execution function
    New() (*cobra.Command, CommandExecFunc)
}
```

**Implementation Requirements:**
- `Tools()`: Must return a stable list of tools derived from your command tree
- `New()`: Must create completely fresh command instances on each call to prevent state pollution

### Config Structure

Configuration for the MCP server bridge:

```go
type Config struct {
    AppName    string // Required: Application name for identification
    AppVersion string // Application version (default: "unknown")
    LogFile    string // Custom log file path (default: user cache)
    LogLevel   string // Log level: debug, info, warn, error (default: "info")
}
```

### Type Mapping

Ophis automatically maps Cobra flag types to MCP parameter types:

| Cobra Type | MCP Type | Description |
|------------|----------|-------------|
| `string` | `string` | Text values |
| `int`, `int8`, `int16`, `int32`, `int64` | `integer` | Numeric values |
| `uint`, `uint8`, `uint16`, `uint32`, `uint64` | `integer` | Unsigned numeric values |
| `float32`, `float64` | `number` | Floating point values |
| `bool` | `boolean` | True/false values |
| `stringSlice`, `stringArray` | `stringArray` | Multiple text values |
| `intSlice` | `intArray` | Multiple numeric values |
| `duration` | `string` | Time duration strings |

## Project Structure

```
ophis/
├── bridge/           # Core MCP server bridge logic
│   ├── config.go     # Server configuration and logging
│   ├── execution.go  # Command execution logic
│   ├── manager.go    # MCP server manager
│   └── registration.go # Tool registration
├── tools/            # Command-to-tool conversion
│   ├── command.go    # Cobra command to MCP tool conversion
│   └── tool.go       # MCP tool definitions
├── mcp/              # Built-in MCP commands
│   ├── claude/       # Claude Desktop integration
│   │   ├── config/   # Config file management
│   │   ├── enable.go # Server enable command
│   │   ├── disable.go# Server disable command
│   │   └── list.go   # Server list command
│   ├── root.go       # Main MCP command
│   ├── start.go      # Server start command
│   └── tools.go      # Tool listing command
└── examples/         # Example implementations
    └── make/         # Make build system example
```

## Examples

### Basic Make Tool

The included example demonstrates exposing `make` commands as MCP tools:

```go
// Create command factory
factory := &CommandFactory{
    rootCmd: createMakeCommands(),
}

// Configure MCP server
config := &bridge.Config{
    AppName:    "make",
    AppVersion: "0.0.1",
}

// Add MCP command to your CLI
rootCmd.AddCommand(mcp.Command(factory, config))
```

This exposes commands like:
- `make test` → `make_test` MCP tool
- `make lint` → `make_lint` MCP tool

### Custom CLI Tool

For your own CLI application:

```go
type MyCommandFactory struct {
    rootCmd *cobra.Command
}

func (f *MyCommandFactory) Tools() []tools.Tool {
    return tools.FromRootCmd(f.rootCmd)
}

func (f *MyCommandFactory) New() (*cobra.Command, bridge.CommandExecFunc) {
    var output strings.Builder
    cmd := createYourCommand() // Your existing command creation
    cmd.SetOut(&output)
    cmd.SetErr(&output)

    execFunc := func(ctx context.Context, cmd *cobra.Command) *mcp.CallToolResult {
        err := cmd.ExecuteContext(ctx)
        if err != nil {
            return mcp.NewToolResultErrorFromErr("Command failed", err)
        }
        return mcp.NewToolResultText(output.String())
    }

    return cmd, execFunc
}
```

## Advanced Usage

### Custom Command Execution

You can customize how commands are executed by implementing a custom `CommandExecFunc`:

```go
execFunc := func(ctx context.Context, cmd *cobra.Command) *mcp.CallToolResult {
    // Pre-execution logic
    startTime := time.Now()
    
    // Execute command
    err := cmd.ExecuteContext(ctx)
    
    // Post-execution logic
    duration := time.Since(startTime)
    
    if err != nil {
        return mcp.NewToolResultError(fmt.Sprintf("Failed after %v: %s", duration, err))
    }
    
    result := fmt.Sprintf("Success (took %v):\n%s", duration, output.String())
    return mcp.NewToolResultText(result)
}
```

### Error Handling

Ophis provides robust error handling:

- **Command Not Found**: Returns descriptive error when command path is invalid
- **Flag Validation**: Cobra's built-in flag validation is preserved
- **Execution Errors**: Command failures are captured and returned as MCP errors
- **Panic Recovery**: Panics during command execution are caught and handled gracefully

### Logging Configuration

Configure detailed logging for debugging:

```bash
# Enable debug logging to file
./your-cli mcp start --log-level debug --log-file /tmp/mcp-debug.log

# Or set via Claude enable command
./your-cli mcp claude enable --log-level debug --log-file /tmp/mcp-debug.log
```

Log files include:
- Tool registration details
- Parameter mapping information
- Command execution traces
- Error details and stack traces

## Security Considerations

- **Command Isolation**: Each tool execution runs in isolation without shell access
- **No Shell Injection**: Commands are executed directly, not through shell interpretation
- **Controlled Exposure**: Only explicitly registered commands are available as MCP tools
- **Parameter Validation**: All parameters go through Cobra's built-in validation
- **Fresh State**: Each execution uses a new command instance, preventing state leakage
- **Access Control**: MCP servers only expose tools, not arbitrary command execution

## Troubleshooting

### Common Issues

**1. Command not found in MCP client**
- Verify command has a `Run` or `RunE` function
- Check that command is not hidden (`Hidden: true`)
- Ensure command is not named `mcp` (reserved)

**2. Flags not working properly**
- Check flag type mapping in the type table above
- Verify flag is not hidden in Cobra
- Test flag locally with Cobra first

**3. Claude Desktop not seeing server**
- Restart Claude Desktop after enabling
- Check config file location with `mcp claude list`
- Verify executable path is correct and accessible

**4. Logging issues**
- Check log file permissions
- Verify log directory exists and is writable
- Use absolute paths for log files

### Debug Commands

```bash
# Export tool definitions for inspection
./your-cli mcp tools

# List Claude configuration
./your-cli mcp claude list

# Test command execution locally
./your-cli your-command --help
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup and contribution guidelines.

### Development Commands

```bash
# Install dependencies
make up

# Run tests
make test

# Run linter
make lint

# Build examples
make build
```

## License

Licensed under the Apache License 2.0. See the [LICENSE](LICENSE) file for details.

## Related Projects

- [Model Context Protocol (MCP)](https://modelcontextprotocol.io/) - The protocol specification
- [MCP-Go Server](https://github.com/mark3labs/mcp-go) - The MCP golang package used by Ophis
- [Cobra](https://github.com/spf13/cobra) - CLI framework for Go
- [Claude Desktop](https://claude.ai/download) - AI assistant with MCP support
- [MCP Servers](https://github.com/modelcontextprotocol/servers) - Collection of MCP server implementations