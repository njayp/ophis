// Package claude provides internal CLI commands for managing Claude Desktop MCP server configuration.
//
// This package implements the 'mcp claude' subcommands that allow users to:
//   - Enable MCP servers in Claude Desktop configuration
//   - Disable MCP servers from Claude Desktop configuration
//   - List currently enabled MCP servers in Claude Desktop
//
// The package manages the claude_desktop_config.json file across different operating systems,
// handling platform-specific configuration paths and formats.
//
// This is an internal package and should not be imported directly by users of the ophis library.
// These commands are automatically available when using ophis.Command() in your application.
package claude
