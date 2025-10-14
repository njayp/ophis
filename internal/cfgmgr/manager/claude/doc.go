// Package claude provides configuration management for Claude Desktop MCP servers.
//
// This package handles:
//   - Claude Desktop configuration structure (claude_desktop_config.json)
//   - Platform-specific configuration file paths (macOS, Linux, Windows)
//   - MCP server entry management
//
// Platform-specific path functions use build tags to locate the Claude Desktop
// configuration directory on different operating systems.
//
// This is an internal package and should not be imported by users of the ophis library.
package claude
