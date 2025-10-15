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
// In addition to tools, Ophis can register MCP resources, resource templates,
// and prompts using the Config.Prompts, Config.Resources, and
// Config.ResourceTemplates fields.
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
//	            FlagSelector: ophis.AllowFlags("namespace", "output"),
//	            // Optional: Add middleware hooks
//	            PreRun: func(ctx context.Context, req *mcp.CallToolRequest, in bridge.ToolInput) (context.Context, *mcp.CallToolRequest, bridge.ToolInput) {
//	                // Add timeout, logging, auth checks, etc.
//	                return ctx, req, in
//	            },
//	            PostRun: func(ctx context.Context, req *mcp.CallToolRequest, in bridge.ToolInput, res *mcp.CallToolResult, out bridge.ToolOutput, err error) (*mcp.CallToolResult, bridge.ToolOutput, error) {
//	                // Error handling, response filtering, metrics
//	                return res, out, err
//	            },
//	        },
//	        {
//	            CmdSelector: ophis.AllowCmds("mycli delete"),
//	            FlagSelector: ophis.ExcludeFlags("all", "force"),
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
// rules and middleware hooks, enabling precise control over the exposed tool surface.
// Each selector defines which commands to match, which flags to include, and optional
// PreRun/PostRun hooks for middleware functionality like timeouts, logging, and filtering.
package ophis
