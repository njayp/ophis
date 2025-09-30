package bridge

import (
	"encoding/json"
	"fmt"

	"github.com/google/jsonschema-go/jsonschema"
)

var (
	inputSchema  = newSchemaCache[CmdToolInput]()
	outputSchema = newSchemaCache[CmdToolOutput]()
)

// CmdToolInput represents the input structure for command tools.
// Do not `omitempty` the Flags field, it helps the AI.
type CmdToolInput struct {
	Flags map[string]any `json:"flags" jsonschema:"Command line flags"`
	Args  []string       `json:"args,omitempty" jsonschema:"Positional command line arguments"`
}

// CmdToolOutput represents the output structure for command tools.
type CmdToolOutput struct {
	StdOut   string `json:"stdout,omitempty" jsonschema:"Standard output"`
	StdErr   string `json:"stderr,omitempty" jsonschema:"Standard error"`
	ExitCode int    `json:"exitCode" jsonschema:"Exit code"`
}

func newSchemaCache[T any]() *schemaCache {
	schema, err := jsonschema.For[T](nil)
	if err != nil {
		panic(fmt.Sprintf("Failed to generate schema: %v", err))
	}

	data, err := json.Marshal(schema)
	if err != nil {
		panic(fmt.Sprintf("Failed to marshal schema: %v", err))
	}

	return &schemaCache{
		data: data,
	}
}

type schemaCache struct {
	data []byte
}

func (s *schemaCache) copy() *jsonschema.Schema {
	schema := &jsonschema.Schema{}
	err := json.Unmarshal(s.data, schema)
	if err != nil {
		panic(fmt.Sprintf("Failed to unmarshal schema: %v", err))
	}

	return schema
}
