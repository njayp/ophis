package bridge

// CmdToolInput represents the input structure for command tools.
type CmdToolInput struct {
	Flags map[string]any `json:"flags,omitempty" jsonschema:"Command line flags"`
	Args  []string       `json:"args,omitempty" jsonschema:"Positional command line arguments"`
}

// CmdToolOutput represents the output structure for command tools.
type CmdToolOutput struct {
	StdOut   string `json:"stdout,omitempty" jsonschema:"Standard output"`
	StdErr   string `json:"stderr,omitempty" jsonschema:"Standard error"`
	ExitCode int    `json:"exitCode" jsonschema:"Exit code"`
}
