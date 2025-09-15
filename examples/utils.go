package examples

import (
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/xeipuuv/gojsonschema"
)

//go:embed schema.json
var schemaData []byte

// ValidateToolAgainstMCP validates a tool definition against the MCP schema.
func ValidateToolAgainstMCP(tool map[string]any) error {
	var schema map[string]interface{}
	if err := json.Unmarshal(schemaData, &schema); err != nil {
		return err
	}

	// Extract Tool definition and create validator
	toolDef := schema["definitions"].(map[string]interface{})["Tool"]
	toolSchema := map[string]interface{}{
		"$schema":     "http://json-schema.org/draft-07/schema#",
		"definitions": schema["definitions"],
	}
	for k, v := range toolDef.(map[string]interface{}) {
		toolSchema[k] = v
	}

	schemaLoader := gojsonschema.NewGoLoader(toolSchema)
	toolLoader := gojsonschema.NewGoLoader(tool)

	result, err := gojsonschema.Validate(schemaLoader, toolLoader)
	if err != nil {
		return err
	}

	if !result.Valid() {
		return fmt.Errorf("validation failed: %v", result.Errors())
	}
	return nil
}
