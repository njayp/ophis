// Package cfgmgr provides internal configuration management utilities for MCP server configurations
// across different platforms (Claude Desktop, VSCode, etc.).
//
// This package contains shared functionality used by the claude and vscode subpackages
// to manage their respective configuration files. It handles common operations like
// validating executables, deriving server names, and other cross-platform utilities.
//
// This is an internal package and should not be imported by users of the ophis library.
package cfgmgr
