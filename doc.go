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
// # Configuration
//
// The Config struct provides fine-grained control over which commands and flags
// are exposed as MCP tools through a powerful selector system.
//
// Basic filters are always applied automatically:
//   - Hidden and deprecated commands/flags are excluded
//   - Commands without executable functions are excluded
//   - Built-in commands (mcp, help, completion) are excluded
//
// Your selectors add additional filtering on top:
//
//	config := &ophis.Config{
//	    // Selectors are evaluated in order - first match wins
//	    Selectors: []ophis.Selector{
//	        {
//	            CmdSelector: ophis.AllowCmdsContaining("get", "list"),
//	            // Control which flags are included for matched commands
//	            LocalFlagSelector: ophis.AllowFlags("namespace", "output"),
//	            InheritedFlagSelector: ophis.NoFlags,  // Exclude persistent flags
//	            // Optional: Add middleware to wrap execution
//	            Middleware: func(ctx context.Context, req *mcp.CallToolRequest, in ophis.ToolInput, next func(context.Context, *mcp.CallToolRequest, ophis.ToolInput) (*mcp.CallToolResult, ophis.ToolOutput, error)) (*mcp.CallToolResult, ophis.ToolOutput, error) {
//	                // Pre-execution: timeout, logging, auth checks, etc.
//	                ctx, cancel := context.WithTimeout(ctx, time.Minute)
//	                defer cancel()
//	                // Execute the command
//	                res, out, err := next(ctx, req, in)
//	                // Post-execution: error handling, response filtering, metrics
//	                return res, out, err
//	            },
//	        },
//	        {
//	            CmdSelector: ophis.AllowCmds("mycli delete"),
//	            LocalFlagSelector: ophis.ExcludeFlags("all", "force"),
//	            InheritedFlagSelector: ophis.NoFlags,
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
// rules and middleware, enabling precise control over the exposed tool surface.
// Each selector defines which commands to match, which flags to include, and optional
// middleware for wrapping execution with timeouts, logging, and response filtering.
package ophis
