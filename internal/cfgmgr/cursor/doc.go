// Package cursor provides internal CLI commands for managing Cursor MCP server configuration.
//
// This package implements the 'mcp cursor' subcommands that allow users to:
//   - Enable MCP servers in Cursor configuration
//   - Disable MCP servers from Cursor configuration
//   - List currently enabled MCP servers in Cursor
//
// The package manages both workspace-level (.cursor/mcp.json) and user-level (user settings)
// MCP configurations, with support for the Cursor MCP extension.
//
// This is an internal package and should not be imported directly by users of the ophis library.
// These commands are automatically available when using ophis.Command() in your application.
package cursor
