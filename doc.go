// Package ophis transforms Cobra CLI applications into MCP (Model Context Protocol) servers,
// enabling AI assistants to interact with command-line tools.
//
// Ophis automatically converts existing Cobra commands into MCP tools, handling
// protocol complexity, command execution, and tool registration.
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
//	    GeneratorOptions: []tools.GeneratorOption{
//	        tools.WithFilters(tools.Allow([]string{"get", "list"})),
//	        tools.WithHandler(customHandler),
//	    },
//
//	    // Configure logging
//	    SloggerOptions: &slog.HandlerOptions{
//	        Level: slog.LevelDebug,
//	    },
//	}
//
//	rootCmd.AddCommand(ophis.Command(config))
//
// # Integration
//
// Enable MCP support in Claude Desktop or VSCode:
//
//	# Claude Desktop
//	./my-cli mcp claude enable
//
//	# VSCode (requires Copilot in Agent Mode)
//	./my-cli mcp vscode enable
//
// Restart the application to load the configuration.
//
// # Architecture
//
// Ophis provides two public packages:
//   - ophis: Main package with Command() function and Config struct
//   - ophis/tools: Filtering and customization utilities
//
// Internal packages contain implementation details and are not part of the public API.
package ophis
