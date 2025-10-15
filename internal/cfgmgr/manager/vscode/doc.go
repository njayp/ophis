// Package vscode provides configuration management for VSCode MCP servers.
//
// This package handles:
//   - VSCode workspace configuration (.vscode/mcp.json)
//   - VSCode user-level configuration
//   - Platform-specific configuration file paths (macOS, Linux, Windows)
//   - MCP server entry management
//
// Platform-specific path functions use build tags to locate VSCode
// configuration directories on different operating systems.
//
// This is an internal package and should not be imported by users of the ophis library.
package vscode
