// Package config provides internal configuration structures and platform-specific
// implementations for managing VSCode's MCP server configuration files.
//
// This package handles:
//   - Reading and writing .vscode/mcp.json files for workspace configuration
//   - Managing user-level VSCode settings for MCP servers
//   - Platform-specific VSCode configuration paths (macOS, Linux, Windows)
//   - Configuration file structure and JSON serialization
//   - MCP server entry management within VSCode settings
//
// The package uses build tags to provide platform-specific implementations for
// locating VSCode configuration directories on different operating systems.
//
// This is an internal package and should not be imported by users of the ophis library.
package vscode
