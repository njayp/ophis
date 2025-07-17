![Project Logo](./logo.png)

**Transform any Cobra CLI application into an MCP (Model Context Protocol) server**

## How it Works

- **Command Tree Traversal**: The `tools.FromRootCmd()` function recursively walks through your `cobra.Command` tree
- **Metadata Extraction**: For each command, command descriptions, flags, and usage information are captured and converted to an `mcp.Tool`
- **Bridge Server**: An mcp server converts tool calls back into `cobra.Commands`, and runs them
- **Prebuilt Commands**: Prebuilt commands install the server into Claude and start the server

## Quick Start

### 1. Add the MCP Command.

Add MCP server capability to your existing Cobra application:

```go
package main

import (
    "github.com/njayp/ophis/bridge"
    "github.com/njayp/ophis/mcp"
    "github.com/spf13/cobra"
)

func main() {
    // You can also add mcp.Command to your root cmd inside createYourExistingCommand
    rootCmd := createYourExistingCommand()
    
    // Create MCP server config
    config := &bridge.Config{
        AppName:    "my-cli-server",
        AppVersion: "1.0.0",
        RootCmd:    rootCmd,
    }
    
    rootCmd.AddCommand(mcp.Command(config))
    
    if err := rootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}
```

### 2. Enable in Claude Desktop

Once your application is built, enable it as an MCP server:

```bash
# Enable your CLI as an MCP server in Claude Desktop
./your-cli mcp claude enable
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md)