package bridge

// CmdToolInput represents the input structure for command tools.
type CmdToolInput struct {
	Flags map[string]interface{} `json:"flags" jsonschema:"Command line flags"`
	Args  []string               `json:"args" jsonschema:"Positional command line arguments"`
}

// CmdToolOutput represents the output structure for command tools.
type CmdToolOutput struct {
	StdOut   string `json:"stdout" jsonschema:"Standard output"`
	StdErr   string `json:"stderr,omitempty" jsonschema:"Standard error"`
	ExitCode int    `json:"exitCode" jsonschema:"Command exit code"`
}
