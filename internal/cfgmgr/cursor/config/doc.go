// Package config provides internal configuration structures and platform-specific
// implementations for managing Cursor's MCP server configuration files.
//
// This package handles:
//   - Reading and writing .cursor/mcp.json files for workspace configuration
//   - Managing user-level Cursor settings for MCP servers
//   - Platform-specific Cursor configuration paths (macOS, Linux, Windows)
//   - Configuration file structure and JSON serialization
//   - MCP server entry management within Cursor settings
//
// The package uses build tags to provide platform-specific implementations for
// locating Cursor configuration directories on different operating systems.
//
// This is an internal package and should not be imported by users of the ophis library.
package config
