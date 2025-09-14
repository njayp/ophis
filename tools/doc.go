// Package tools converts Cobra commands into MCP tools with customizable filtering and output handling.
//
// # Core Components
//
// Generator: Converts Cobra command trees to MCP tools.
//
//	generator := tools.NewGenerator(
//	    tools.WithFilters(filters...),  // Control which commands are exposed
//	    tools.WithHandler(handler),     // Customize output processing
//	)
//
// Filters: Control which commands become tools.
//   - Allow([]string): Only expose specified commands
//   - Exclude([]string): Hide specified commands
//   - Hidden(): Skip hidden commands (default)
//   - Runs(): Skip non-runnable commands (default)
//
// Handlers: Process command output before returning to MCP clients.
// Default handler returns plain text. Custom handlers can format as JSON, images, etc.
//
// # Example Usage
//
//	config := &ophis.Config{
//	    GeneratorOptions: []tools.GeneratorOption{
//	        // Only expose safe commands
//	        tools.WithFilters(tools.Allow([]string{"get", "list"})),
//
//	        // Custom output handler
//	        tools.WithHandler(customHandler),
//	    },
//	}
package tools
