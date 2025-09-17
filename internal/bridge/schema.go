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
