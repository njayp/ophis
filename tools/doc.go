// Package tools provides utilities for converting Cobra commands into MCP tools.
//
// This package offers customization options for controlling how Cobra commands are
// exposed as MCP tools, including filtering mechanisms and output handlers.
//
// # Core Components
//
// The package provides several key components for customization:
//
// Generator: Controls the conversion process from Cobra commands to MCP tools.
// Use NewGenerator() with various options to customize the behavior:
//
//	generator := tools.NewGenerator(
//	    tools.WithFilters(filters...),  // Control which commands are exposed
//	    tools.WithHandler(handler),     // Customize output processing
//	)
//
// Filters: Functions that determine which commands should be exposed as tools.
// Built-in filters include:
//   - Allow([]string): Only expose commands with specific names
//   - Exclude([]string): Hide specific commands from MCP
//   - Hidden(): Exclude hidden Cobra commands (applied by default)
//
// Handlers: Functions that process command output before returning it to MCP clients.
// The default handler returns output as plain text, but you can provide custom
// handlers to format output differently (e.g., as JSON, images, or structured data).
//
// # Example Usage
//
//	config := &ophis.Config{
//	    Generator: tools.NewGenerator(
//	        // Only expose safe read-only commands
//	        tools.WithFilters(tools.Allow([]string{"get", "list", "describe"})),
//
//	        // Exclude potentially dangerous commands
//	        tools.AddFilter(tools.Exclude([]string{"delete", "destroy"})),
//
//	        // Custom output processing
//	        tools.WithHandler(func(ctx context.Context, req mcp.CallToolRequest, data []byte) *mcp.CallToolResult {
//	            // Process the output data
//	            return mcp.NewToolResultText(string(data))
//	        }),
//	    ),
//	}
//
// # Default Behavior
//
// When no custom Generator is provided, the default behavior:
//   - Excludes hidden commands
//   - Excludes "mcp", "help", and "completion" commands
//   - Returns command output as plain text
//
// This package is part of the public API and can be imported by users who need
// fine-grained control over the command-to-tool conversion process.
package tools
