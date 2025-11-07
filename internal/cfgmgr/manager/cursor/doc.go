// Package cursor provides configuration management for Cursor MCP servers.
//
// This package handles:
//   - Cursor workspace configuration (.cursor/mcp.json)
//   - Cursor user-level configuration
//   - Platform-specific configuration file paths (macOS, Linux, Windows)
//   - MCP server entry management
//
// Platform-specific path functions use build tags to locate Cursor
// configuration directories on different operating systems.
//
// This is an internal package and should not be imported by users of the ophis library.
package cursor
