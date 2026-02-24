// Package zed provides CLI commands for managing Zed MCP context servers.
//
// This package implements the 'mcp zed' subcommands:
//   - enable: Add MCP server to Zed configuration
//   - disable: Remove MCP server from Zed configuration
//   - list: Show all configured MCP servers
//
// Supports both workspace (.zed/settings.json) and user-level configurations
// (~/.config/zed/settings.json). When modifying the settings file, all existing
// Zed settings (theme, fonts, keybindings, etc.) are preserved unchanged.
//
// This is an internal package and should not be imported directly by users of the ophis library.
// These commands are automatically available when using ophis.Command() in your application.
package zed
