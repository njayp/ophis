package bridge

import (
	"encoding/json"
	"fmt"

	"github.com/google/jsonschema-go/jsonschema"
)

var (
	inputSchema  = newSchemaCache[ToolInput]()
	outputSchema = newSchemaCache[ToolOutput]()
)

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
