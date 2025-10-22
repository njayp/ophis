package schema

import (
	"encoding/json"
	"fmt"

	"github.com/google/jsonschema-go/jsonschema"
)

// New creates a new schema cache for the given type.
func New[T any]() *Cache {
	schema, err := jsonschema.For[T](nil)
	if err != nil {
		panic(fmt.Sprintf("Failed to generate schema: %v", err))
	}

	data, err := json.Marshal(schema)
	if err != nil {
		panic(fmt.Sprintf("Failed to marshal schema: %v", err))
	}

	return &Cache{
		data: data,
	}
}

// Cache stores a cached JSON schema.
type Cache struct {
	data []byte
}

// Copy returns a copy of the cached schema.
func (s *Cache) Copy() *jsonschema.Schema {
	schema := &jsonschema.Schema{}
	err := json.Unmarshal(s.data, schema)
	if err != nil {
		panic(fmt.Sprintf("Failed to unmarshal schema: %v", err))
	}

	return schema
}
