// Package vscode provides internal CLI commands for managing VSCode MCP server configuration.
//
// This package implements the 'mcp vscode' subcommands that allow users to:
//   - Enable MCP servers in VSCode configuration
//   - Disable MCP servers from VSCode configuration
//   - List currently enabled MCP servers in VSCode
//
// The package manages both workspace-level (.vscode/mcp.json) and user-level (user settings)
// MCP configurations, with support for the VSCode MCP extension.
//
// This is an internal package and should not be imported directly by users of the ophis library.
// These commands are automatically available when using ophis.Command() in your application.
package vscode
