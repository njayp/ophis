package bridge

// CmdToolInput represents the input structure for command tools.
// Do not omitempty -- or ai will "forget" to send these parameters even if
// it displays them for the user
type CmdToolInput struct {
	Flags map[string]any `json:"flags" jsonschema:"Command line flags"`
	Args  []string       `json:"args" jsonschema:"Positional command line arguments"`
}

// CmdToolOutput represents the output structure for command tools.
type CmdToolOutput struct {
	StdOut   string `json:"stdout,omitempty" jsonschema:"Standard output"`
	StdErr   string `json:"stderr,omitempty" jsonschema:"Standard error"`
	ExitCode int    `json:"exitCode" jsonschema:"Exit code"`
}
