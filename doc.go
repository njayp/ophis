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
// The Config struct provides fine-grained control over which commands and flags
// are exposed as MCP tools through a powerful selector system:
//
//	config := &ophis.Config{
//	    // Selectors are evaluated in order - first match wins
//	    Selectors: []ophis.Selector{
//	        {
//	            // Match specific commands
//	            CmdSelect: ophis.AllowCmd("get", "list"),
//	            // Control which flags are included for matched commands
//	            FlagSelect: ophis.AllowFlag("namespace", "output"),
//	        },
//	        {
//	            // Different flag rules for different commands
//	            CmdSelect: ophis.AllowCmd("delete"),
//	            FlagSelect: ophis.ExcludeFlag("all", "force"),
//	        },
//	    },
//
//	    // Configure logging
//	    SloggerOptions: &slog.HandlerOptions{
//	        Level: slog.LevelDebug,
//	    },
//	}
//
// The selector system allows different commands to have different flag filtering
// rules, enabling precise control over the exposed tool surface. Each selector
// defines both which commands to match and which of their flags to include.
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
package ophis
