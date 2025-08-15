// Package config provides internal configuration structures and platform-specific
// implementations for managing Claude Desktop's MCP server configuration files.
//
// This package handles:
//   - Reading and writing claude_desktop_config.json files
//   - Platform-specific configuration file locations (macOS, Linux, Windows)
//   - Configuration file structure and JSON serialization
//   - MCP server entry management within the configuration
//
// The package uses build tags to provide platform-specific implementations for
// locating the Claude Desktop configuration directory on different operating systems.
//
// This is an internal package and should not be imported by users of the ophis library.
package config
