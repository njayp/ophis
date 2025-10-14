// Package vscode provides CLI commands for managing VSCode MCP servers.
//
// This package implements the 'mcp vscode' subcommands:
//   - enable: Add MCP server to VSCode configuration
//   - disable: Remove MCP server from VSCode configuration
//   - list: Show all configured MCP servers
//
// Supports both workspace (.vscode/mcp.json) and user-level configurations.
//
// This is an internal package and should not be imported directly by users of the ophis library.
// These commands are automatically available when using ophis.Command() in your application.
package vscode
