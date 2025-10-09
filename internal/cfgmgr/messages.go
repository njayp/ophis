package cfgmgr

// Common user-facing messages
const (
	// MsgRestartClaudeDesktop is shown after config changes for Claude Desktop
	MsgRestartClaudeDesktop = "\nTo use this server, restart Claude Desktop."

	// MsgRestartVSCode is shown after config changes for VSCode
	MsgRestartVSCode = "\nTo apply changes, restart VSCode or reload the window."

	// MsgServerOverwrite warns when a server will be overwritten
	MsgServerOverwrite = "⚠️  MCP server %q already exists and will be overwritten\n"

	// MsgBackupCreated confirms backup creation
	MsgBackupCreated = "Backup config file created at %q\n"

	// MsgServerEnabled confirms server was enabled
	MsgServerEnabled = "Successfully enabled MCP server %q\n"

	// MsgServerDisabled confirms server was disabled
	MsgServerDisabled = "Successfully disabled MCP server %q\n"

	// MsgServerNotEnabled informs that server is not enabled
	MsgServerNotEnabled = "MCP server %q is not currently enabled\n"

	// MsgNoServersConfigured indicates no servers are set up
	MsgNoServersConfigured = "No MCP servers are currently configured."
)
