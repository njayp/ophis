package bridge

import (
	"encoding/json"
	"fmt"

	"github.com/google/jsonschema-go/jsonschema"
)

var (
	inputSchemaBytes []byte
	outputSchema     *jsonschema.Schema
)

func initInputSchemaBytes() {
	schema, err := jsonschema.For[CmdToolInput](nil)
	if err != nil {
		panic(fmt.Sprintf("Failed to generate input schema: %v", err))
	}

	data, err := json.Marshal(schema)
	if err != nil {
		panic(fmt.Sprintf("Failed to marshal input schema: %v", err))
	}

	inputSchemaBytes = data
}

func initOutputSchema() {
	schema, err := jsonschema.For[CmdToolOutput](nil)
	if err != nil {
		panic(fmt.Sprintf("Failed to generate output schema: %v", err))
	}
	outputSchema = schema
}

func newInputSchema() *jsonschema.Schema {
	schema := &jsonschema.Schema{}
	err := json.Unmarshal(inputSchemaBytes, schema)
	if err != nil {
		panic(fmt.Sprintf("Failed to unmarshal input schema: %v", err))
	}

	return schema
}
