package bridge

import (
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseArrayDefault(t *testing.T) {
	tests := []struct {
		name       string
		defValue   string
		itemSchema *jsonschema.Schema
		expected   any
	}{
		{
			name:       "empty array brackets",
			defValue:   "[]",
			itemSchema: &jsonschema.Schema{Type: "string"},
			expected:   nil,
		},
		{
			name:       "empty inner content",
			defValue:   "[]",
			itemSchema: &jsonschema.Schema{Type: "integer"},
			expected:   nil,
		},
		{
			name:       "invalid format - no brackets",
			defValue:   "item1,item2",
			itemSchema: &jsonschema.Schema{Type: "string"},
			expected:   nil,
		},
		{
			name:       "invalid format - only opening bracket",
			defValue:   "[item1,item2",
			itemSchema: &jsonschema.Schema{Type: "string"},
			expected:   nil,
		},
		{
			name:       "invalid format - only closing bracket",
			defValue:   "item1,item2]",
			itemSchema: &jsonschema.Schema{Type: "string"},
			expected:   nil,
		},
		{
			name:       "string array",
			defValue:   "[hello,world,test]",
			itemSchema: &jsonschema.Schema{Type: "string"},
			expected:   []string{"hello", "world", "test"},
		},
		{
			name:       "string array with spaces",
			defValue:   "[ hello , world , test ]",
			itemSchema: &jsonschema.Schema{Type: "string"},
			expected:   []string{"hello", "world", "test"},
		},
		{
			name:       "integer array",
			defValue:   "[1,2,3]",
			itemSchema: &jsonschema.Schema{Type: "integer"},
			expected:   []int64{1, 2, 3},
		},
		{
			name:       "integer array with spaces",
			defValue:   "[ 10 , 20 , 30 ]",
			itemSchema: &jsonschema.Schema{Type: "integer"},
			expected:   []int64{10, 20, 30},
		},
		{
			name:       "integer array with invalid values",
			defValue:   "[1,invalid,3]",
			itemSchema: &jsonschema.Schema{Type: "integer"},
			expected:   []int64{1, 3},
		},
		{
			name:       "float array",
			defValue:   "[1.5,2.7,3.14]",
			itemSchema: &jsonschema.Schema{Type: "number"},
			expected:   []float64{1.5, 2.7, 3.14},
		},
		{
			name:       "float array with spaces",
			defValue:   "[ 1.5 , 2.7 , 3.14 ]",
			itemSchema: &jsonschema.Schema{Type: "number"},
			expected:   []float64{1.5, 2.7, 3.14},
		},
		{
			name:       "boolean array",
			defValue:   "[true,false,true]",
			itemSchema: &jsonschema.Schema{Type: "boolean"},
			expected:   []bool{true, false, true},
		},
		{
			name:       "boolean array with spaces",
			defValue:   "[ true , false , true ]",
			itemSchema: &jsonschema.Schema{Type: "boolean"},
			expected:   []bool{true, false, true},
		},
		{
			name:       "unknown type",
			defValue:   "[item1,item2]",
			itemSchema: &jsonschema.Schema{Type: "unknown"},
			expected:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseArray(tt.defValue, tt.itemSchema)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseIntArray(t *testing.T) {
	tests := []struct {
		name     string
		parts    []string
		expected []int64
	}{
		{
			name:     "valid integers",
			parts:    []string{"1", "2", "3"},
			expected: []int64{1, 2, 3},
		},
		{
			name:     "integers with whitespace",
			parts:    []string{" 1 ", " 2 ", " 3 "},
			expected: []int64{1, 2, 3},
		},
		{
			name:     "mixed valid and invalid",
			parts:    []string{"1", "invalid", "3"},
			expected: []int64{1, 3},
		},
		{
			name:     "negative integers",
			parts:    []string{"-1", "-2", "-3"},
			expected: []int64{-1, -2, -3},
		},
		{
			name:     "large integers",
			parts:    []string{"9223372036854775807"},
			expected: []int64{9223372036854775807},
		},
		{
			name:     "empty array",
			parts:    []string{},
			expected: []int64{},
		},
		{
			name:     "all invalid",
			parts:    []string{"abc", "def", "ghi"},
			expected: []int64{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseIntArray(tt.parts)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseFloatArray(t *testing.T) {
	tests := []struct {
		name     string
		parts    []string
		expected []float64
	}{
		{
			name:     "valid floats",
			parts:    []string{"1.5", "2.7", "3.14"},
			expected: []float64{1.5, 2.7, 3.14},
		},
		{
			name:     "floats with whitespace",
			parts:    []string{" 1.5 ", " 2.7 ", " 3.14 "},
			expected: []float64{1.5, 2.7, 3.14},
		},
		{
			name:     "mixed valid and invalid",
			parts:    []string{"1.5", "invalid", "3.14"},
			expected: []float64{1.5, 3.14},
		},
		{
			name:     "negative floats",
			parts:    []string{"-1.5", "-2.7", "-3.14"},
			expected: []float64{-1.5, -2.7, -3.14},
		},
		{
			name:     "integers as floats",
			parts:    []string{"1", "2", "3"},
			expected: []float64{1.0, 2.0, 3.0},
		},
		{
			name:     "scientific notation",
			parts:    []string{"1e10", "2.5e-3"},
			expected: []float64{1e10, 2.5e-3},
		},
		{
			name:     "empty array",
			parts:    []string{},
			expected: []float64{},
		},
		{
			name:     "all invalid",
			parts:    []string{"abc", "def", "ghi"},
			expected: []float64{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseFloatArray(tt.parts)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseBoolArray(t *testing.T) {
	tests := []struct {
		name     string
		parts    []string
		expected []bool
	}{
		{
			name:     "valid booleans",
			parts:    []string{"true", "false", "true"},
			expected: []bool{true, false, true},
		},
		{
			name:     "booleans with whitespace",
			parts:    []string{" true ", " false ", " true "},
			expected: []bool{true, false, true},
		},
		{
			name:     "mixed valid and invalid",
			parts:    []string{"true", "invalid", "false"},
			expected: []bool{true, false},
		},
		{
			name:     "numeric representations",
			parts:    []string{"1", "0", "1"},
			expected: []bool{true, false, true},
		},
		{
			name:  "case variations",
			parts: []string{"True", "FALSE", "tRuE"},
			// tRue is not a valid bool, so it will be ignored
			expected: []bool{true, false},
		},
		{
			name:     "empty array",
			parts:    []string{},
			expected: []bool{},
		},
		{
			name:     "all invalid",
			parts:    []string{"yes", "no", "maybe"},
			expected: []bool{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseBoolArray(tt.parts)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseStringArray(t *testing.T) {
	tests := []struct {
		name     string
		parts    []string
		expected []string
	}{
		{
			name:     "simple strings",
			parts:    []string{"hello", "world", "test"},
			expected: []string{"hello", "world", "test"},
		},
		{
			name:     "strings with whitespace",
			parts:    []string{" hello ", " world ", " test "},
			expected: []string{"hello", "world", "test"},
		},
		{
			name:     "empty strings",
			parts:    []string{"", "", ""},
			expected: []string{"", "", ""},
		},
		{
			name:     "mixed content",
			parts:    []string{"hello", "123", "true", "3.14"},
			expected: []string{"hello", "123", "true", "3.14"},
		},
		{
			name:     "empty array",
			parts:    []string{},
			expected: []string{},
		},
		{
			name:     "strings with special characters",
			parts:    []string{"hello-world", "test_value", "my.file"},
			expected: []string{"hello-world", "test_value", "my.file"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseStringArray(tt.parts)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSetDefaultFromFlag_ArrayTypes(t *testing.T) {
	tests := []struct {
		name              string
		defValue          string
		schemaType        string
		itemType          string
		expectedJSON      string
		shouldHaveDefault bool
	}{
		{
			name:              "string array with values",
			defValue:          "[hello,world]",
			schemaType:        "array",
			itemType:          "string",
			expectedJSON:      `["hello","world"]`,
			shouldHaveDefault: true,
		},
		{
			name:              "integer array with values",
			defValue:          "[1,2,3]",
			schemaType:        "array",
			itemType:          "integer",
			expectedJSON:      `[1,2,3]`,
			shouldHaveDefault: true,
		},
		{
			name:              "float array with values",
			defValue:          "[1.5,2.7,3.14]",
			schemaType:        "array",
			itemType:          "number",
			expectedJSON:      `[1.5,2.7,3.14]`,
			shouldHaveDefault: true,
		},
		{
			name:              "boolean array with values",
			defValue:          "[true,false,true]",
			schemaType:        "array",
			itemType:          "boolean",
			expectedJSON:      `[true,false,true]`,
			shouldHaveDefault: true,
		},
		{
			name:              "empty array",
			defValue:          "[]",
			schemaType:        "array",
			itemType:          "string",
			expectedJSON:      "",
			shouldHaveDefault: false,
		},
		{
			name:              "invalid array format",
			defValue:          "not-an-array",
			schemaType:        "array",
			itemType:          "string",
			expectedJSON:      "",
			shouldHaveDefault: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := &jsonschema.Schema{
				Type: tt.schemaType,
				Items: &jsonschema.Schema{
					Type: tt.itemType,
				},
			}

			// Create a real pflag.Flag with the default value
			flagSet := pflag.NewFlagSet("test", pflag.ContinueOnError)
			switch tt.itemType {
			case "string":
				flagSet.StringSlice("test", []string{}, "test flag")
			case "integer":
				flagSet.IntSlice("test", []int{}, "test flag")
			case "number":
				flagSet.Float64Slice("test", []float64{}, "test flag")
			case "boolean":
				flagSet.BoolSlice("test", []bool{}, "test flag")
			}

			flag := flagSet.Lookup("test")
			require.NotNil(t, flag)
			flag.DefValue = tt.defValue

			setDefaultFromFlag(schema, flag)

			if tt.shouldHaveDefault {
				require.NotNil(t, schema.Default, "Expected default to be set")
				assert.JSONEq(t, tt.expectedJSON, string(schema.Default))
			} else {
				assert.Nil(t, schema.Default, "Expected no default to be set")
			}
		})
	}
}

func TestParseObjectEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		parts    []string
		schema   *jsonschema.Schema
		expected any
	}{
		{
			name:   "string object with whitespace in key and value",
			parts:  []string{" key1 = value1 ", " key2 = value2 "},
			schema: &jsonschema.Schema{Type: "string"},
			expected: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			name:   "integer object with whitespace",
			parts:  []string{" port = 8080 ", " timeout = 30 "},
			schema: &jsonschema.Schema{Type: "integer"},
			expected: map[string]int64{
				"port":    8080,
				"timeout": 30,
			},
		},
		{
			name:     "all malformed string entries returns nil",
			parts:    []string{"invalid", "also-invalid"},
			schema:   &jsonschema.Schema{Type: "string"},
			expected: map[string]string(nil),
		},
		{
			name:     "all malformed int entries returns nil",
			parts:    []string{"invalid", "also-invalid"},
			schema:   &jsonschema.Schema{Type: "integer"},
			expected: map[string]int64(nil),
		},
		{
			name:   "value containing equals sign",
			parts:  []string{"url=https://example.com/path?query=value"},
			schema: &jsonschema.Schema{Type: "string"},
			expected: map[string]string{
				"url": "https://example.com/path?query=value",
			},
		},
		{
			name:   "mixed valid and invalid integer entries",
			parts:  []string{"valid=123", "invalid", "another=456"},
			schema: &jsonschema.Schema{Type: "integer"},
			expected: map[string]int64{
				"valid":   123,
				"another": 456,
			},
		},
		{
			name:   "mixed valid and invalid string entries",
			parts:  []string{"valid=value", "invalid", "another=value2"},
			schema: &jsonschema.Schema{Type: "string"},
			expected: map[string]string{
				"valid":   "value",
				"another": "value2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result any
			switch tt.schema.Type {
			case "string":
				result = parseStringObj(tt.parts)
			case "integer":
				result = parseIntObj(tt.parts)
			}
			if tt.expected == nil || (tt.expected != nil && result == nil) {
				assert.Nil(t, result)
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestSetDefaultFromFlag_ScalarTypes(t *testing.T) {
	tests := []struct {
		name              string
		defValue          string
		schemaType        string
		expectedJSON      string
		shouldHaveDefault bool
	}{
		{
			name:              "boolean true",
			defValue:          "true",
			schemaType:        "boolean",
			expectedJSON:      `true`,
			shouldHaveDefault: true,
		},
		{
			name:              "boolean false",
			defValue:          "false",
			schemaType:        "boolean",
			expectedJSON:      `false`,
			shouldHaveDefault: true,
		},
		{
			name:              "integer",
			defValue:          "42",
			schemaType:        "integer",
			expectedJSON:      `42`,
			shouldHaveDefault: true,
		},
		{
			name:              "negative integer",
			defValue:          "-100",
			schemaType:        "integer",
			expectedJSON:      `-100`,
			shouldHaveDefault: true,
		},
		{
			name:              "float",
			defValue:          "3.14",
			schemaType:        "number",
			expectedJSON:      `3.14`,
			shouldHaveDefault: true,
		},
		{
			name:              "string",
			defValue:          "hello",
			schemaType:        "string",
			expectedJSON:      `"hello"`,
			shouldHaveDefault: true,
		},
		{
			name:              "empty string",
			defValue:          "",
			schemaType:        "string",
			expectedJSON:      "",
			shouldHaveDefault: false,
		},
		{
			name:              "invalid boolean",
			defValue:          "not-a-bool",
			schemaType:        "boolean",
			expectedJSON:      "",
			shouldHaveDefault: false,
		},
		{
			name:              "invalid integer",
			defValue:          "not-an-int",
			schemaType:        "integer",
			expectedJSON:      "",
			shouldHaveDefault: false,
		},
		{
			name:              "invalid float",
			defValue:          "not-a-float",
			schemaType:        "number",
			expectedJSON:      "",
			shouldHaveDefault: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := &jsonschema.Schema{
				Type: tt.schemaType,
			}

			// Create a real pflag.Flag
			flagSet := pflag.NewFlagSet("test", pflag.ContinueOnError)
			flagSet.String("test", "", "test flag")
			flag := flagSet.Lookup("test")
			require.NotNil(t, flag)
			flag.DefValue = tt.defValue

			setDefaultFromFlag(schema, flag)

			if tt.shouldHaveDefault {
				require.NotNil(t, schema.Default, "Expected default to be set")
				assert.JSONEq(t, tt.expectedJSON, string(schema.Default))
			} else {
				assert.Nil(t, schema.Default, "Expected no default to be set")
			}
		})
	}
}
