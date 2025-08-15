// Package ophis transforms Cobra CLI applications into MCP (Model Context Protocol) servers,
// enabling AI assistants like Claude to interact with your command-line tools.
//
// Ophis provides a simple way to expose your existing Cobra commands as MCP tools that can be
// called by AI assistants and other MCP-compatible clients. It handles all the complexity of
// the MCP protocol, command execution, and tool registration.
//
// # Basic Usage
//
// Add MCP server functionality to your existing Cobra CLI application:
//
//	package main
//
//	import (
//	    "os"
//	    "github.com/njayp/ophis"
//	)
//
//	func main() {
//	    rootCmd := createMyRootCommand()
//
//	    // Add MCP server commands
//	    rootCmd.AddCommand(ophis.Command(nil))
//
//	    if err := rootCmd.Execute(); err != nil {
//	        os.Exit(1)
//	    }
//	}
//
// This adds the following subcommands to your CLI:
//   - mcp start: Start the MCP server
//   - mcp tools: List available tools
//   - mcp claude enable/disable/list: Manage Claude Desktop integration
//   - mcp vscode enable/disable/list: Manage VSCode integration
//
// # Configuration
//
// You can customize the MCP server behavior using the Config struct:
//
//	config := &ophis.Config{
//	    // Control which commands are exposed
//	    Generator: tools.NewGenerator(
//	        tools.WithFilters(tools.Allow([]string{"get", "list"})),
//	        tools.WithHandler(customHandler),
//	    ),
//
//	    // Configure logging
//	    SloggerOptions: &slog.HandlerOptions{
//	        Level: slog.LevelDebug,
//	    },
//	}
//
//	rootCmd.AddCommand(ophis.Command(config))
//
// # Integration with AI Assistants
//
// After adding MCP support to your CLI, you can enable it in Claude Desktop or VSCode:
//
//	# Enable in Claude Desktop
//	./my-cli mcp claude enable
//
//	# Enable in VSCode
//	./my-cli mcp vscode enable
//
// Then restart the application to load the MCP server configuration.
//
// # Architecture
//
// Ophis consists of two main public packages:
//   - ophis: The main package providing the Command() function and Config struct
//   - ophis/tools: Utilities for filtering and customizing command-to-tool conversion
//
// All other implementation details are contained in internal packages and are not
// part of the public API. This design ensures API stability while allowing internal
// refactoring and improvements.
package ophis
