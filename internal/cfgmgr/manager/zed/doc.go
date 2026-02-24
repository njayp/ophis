// Package zed provides configuration management for Zed MCP context servers.
//
// This package handles:
//   - Zed workspace configuration (.zed/settings.json)
//   - Zed user-level configuration (~/.config/zed/settings.json)
//   - Platform-specific configuration file paths (macOS, Linux, Windows)
//   - MCP server entry management under the "context_servers" key
//
// Unlike other editors, Zed's settings.json is a general-purpose configuration
// file containing many editor settings beyond MCP servers. This package uses
// custom JSON marshaling to safely read and write only the "context_servers"
// section while preserving all other existing Zed settings unchanged.
//
// Platform-specific path functions use build tags to locate the Zed
// configuration directory on different operating systems. On Linux, the
// XDG_CONFIG_HOME environment variable is respected.
//
// This is an internal package and should not be imported by users of the ophis library.
package zed
