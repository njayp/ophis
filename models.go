package ophis

import "github.com/njayp/ophis/internal/schema"

// ToolInput represents the input structure for command tools.
// Do not `omitempty` the Flags field, there may be required flags inside.
type ToolInput struct {
	Flags map[string]any `json:"flags" jsonschema:"Command line flags"`
	Args  []string       `json:"args,omitempty" jsonschema:"Positional command line arguments"`
}

// ToolOutput represents the output structure for command tools.
type ToolOutput struct {
	StdOut   string `json:"stdout,omitempty" jsonschema:"Standard output"`
	StdErr   string `json:"stderr,omitempty" jsonschema:"Standard error"`
	ExitCode int    `json:"exitCode" jsonschema:"Exit code"`
}

var (
	inputSchema  = schema.New[ToolInput]()
	outputSchema = schema.New[ToolOutput]()
)
