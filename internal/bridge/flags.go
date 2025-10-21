package bridge

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// enhanceFlagsSchema adds detailed flag information to the flags property.
func (s Selector) enhanceFlagsSchema(schema *jsonschema.Schema, cmd *cobra.Command) {
	// Ensure properties map exists
	if schema.Properties == nil {
		schema.Properties = make(map[string]*jsonschema.Schema)
	}

	// Process local flags
	cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
		if s.localFlagSelect(flag) {
			addFlagToSchema(schema, flag)
		}
	})

	// Process inherited flags
	cmd.InheritedFlags().VisitAll(func(flag *pflag.Flag) {
		if s.inheritedFlagSelect(flag) {
			// Skip if already added as local flag
			if _, exists := schema.Properties[flag.Name]; !exists {
				addFlagToSchema(schema, flag)
			}
		}
	})

	// Set AdditionalProperties to false
	// See https://github.com/google/jsonschema-go/issues/13
	schema.AdditionalProperties = &jsonschema.Schema{Not: &jsonschema.Schema{}}
}

// isFlagRequired checks if a flag has been marked as required by Cobra.
// Cobra uses the BashCompOneRequiredFlag annotation to track required flags.
func isFlagRequired(flag *pflag.Flag) bool {
	if flag.Annotations == nil {
		return false
	}

	// Check if the flag has the required annotation
	// The constant is defined as "cobra_annotation_bash_completion_one_required_flag" in Cobra
	if val, ok := flag.Annotations[cobra.BashCompOneRequiredFlag]; ok {
		// The annotation is present if the flag is required
		return len(val) > 0 && val[0] == "true"
	}

	return false
}

// addFlagToSchema adds a single flag to the schema properties.
func addFlagToSchema(schema *jsonschema.Schema, flag *pflag.Flag) {
	flagSchema := &jsonschema.Schema{
		Description: flag.Usage,
	}

	// Check if flag is marked as required in its annotations
	// Cobra uses the BashCompOneRequiredFlag annotation to mark required flags
	if isFlagRequired(flag) {
		// Mark the flag as required in the schema
		if schema.Required == nil {
			schema.Required = []string{}
		}

		schema.Required = append(schema.Required, flag.Name)
	}

	// Set appropriate JSON schema type based on flag type
	t := flag.Value.Type()
	switch t {
	case "bool":
		flagSchema.Type = "boolean"
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "count":
		flagSchema.Type = "integer"
	case "float32", "float64":
		flagSchema.Type = "number"
	case "string":
		flagSchema.Type = "string"
	case "stringSlice", "stringArray":
		flagSchema.Type = "array"
		flagSchema.Items = &jsonschema.Schema{Type: "string"}
	case "intSlice", "int32Slice", "int64Slice", "uintSlice":
		flagSchema.Type = "array"
		flagSchema.Items = &jsonschema.Schema{Type: "integer"}
	case "float32Slice", "float64Slice":
		flagSchema.Type = "array"
		flagSchema.Items = &jsonschema.Schema{Type: "number"}
	case "boolSlice":
		flagSchema.Type = "array"
		flagSchema.Items = &jsonschema.Schema{Type: "boolean"}
	case "stringToString":
		flagSchema.Type = "object"
		flagSchema.AdditionalProperties = &jsonschema.Schema{Type: "string"}
		flagSchema.Description += " (format: key-value pairs)"
	case "stringToInt", "stringToInt64":
		flagSchema.Type = "object"
		flagSchema.AdditionalProperties = &jsonschema.Schema{Type: "integer"}
		flagSchema.Description += " (format: key-value pairs with integer values)"
	case "duration":
		// Duration is represented as a string in Go's duration format
		flagSchema.Type = "string"
		flagSchema.Description += " (format: Go duration string, e.g., '10s', '2h45m')"
		flagSchema.Pattern = `^-?([0-9]+(\.[0-9]+)?(ns|us|µs|ms|s|m|h))+$`
	case "ip":
		flagSchema.Type = "string"
		flagSchema.Description += " (format: IPv4 or IPv6 address)"
		flagSchema.Pattern = `^((25[0-5]|(2[0-4]|1\d|[1-9]|)\d)\.){3}(25[0-5]|(2[0-4]|1\d|[1-9]|)\d)$|^(([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4})$`
	case "ipMask":
		flagSchema.Type = "string"
		flagSchema.Description += " (format: IP mask, e.g., '255.255.255.0')"
	case "ipNet":
		flagSchema.Type = "string"
		flagSchema.Description += " (format: CIDR notation, e.g., '192.168.1.0/24')"
		flagSchema.Pattern = `^((25[0-5]|(2[0-4]|1\d|[1-9]|)\d)\.){3}(25[0-5]|(2[0-4]|1\d|[1-9]|)\d)/([0-9]|[1-2][0-9]|3[0-2])$`
	case "bytesHex":
		flagSchema.Type = "string"
		flagSchema.Description += " (format: hexadecimal string)"
		flagSchema.Pattern = `^[0-9a-fA-F]*$`
	case "bytesBase64":
		flagSchema.Type = "string"
		flagSchema.Description += " (format: base64 encoded string)"
		flagSchema.Pattern = `^[A-Za-z0-9+/]*={0,2}$`
	default:
		// Default to string for unknown types
		flagSchema.Type = "string"
		flagSchema.Description += fmt.Sprintf(" (type: %s)", t)
		slog.Debug("unknown flag type, defaulting to string", "flag", flag.Name, "type", t)
	}

	setDefaultFromFlag(flagSchema, flag)
	schema.Properties[flag.Name] = flagSchema
}

// setDefaultFromFlag sets the default value for a flag schema if it's not a zero value.
func setDefaultFromFlag(flagSchema *jsonschema.Schema, flag *pflag.Flag) {
	defValue := flag.DefValue
	if defValue == "" {
		return
	}

	setDefault := func(val any) {
		if raw, err := json.Marshal(val); err == nil {
			flagSchema.Default = json.RawMessage(raw)
		}
	}

	// Parse the default value based on the schema type
	switch flagSchema.Type {
	case "boolean":
		if val, err := strconv.ParseBool(defValue); err == nil {
			setDefault(val)
		}
	case "integer":
		if val, err := strconv.ParseInt(defValue, 10, 64); err == nil {
			setDefault(val)
		}
	case "number":
		if val, err := strconv.ParseFloat(defValue, 64); err == nil {
			setDefault(val)
		}
	case "string":
		setDefault(defValue)
	case "array":
		if arr := parseArray(defValue, flagSchema.Items); arr != nil {
			setDefault(arr)
		}
	case "object":
		if obj := parseObject(defValue, flagSchema.AdditionalProperties); obj != nil {
			setDefault(obj)
		}
	}
}

// parseArray parses pflag's array representation ("[item1,item2]") into a typed slice.
// Invalid array elements are skipped and logged as warnings. Returns nil for malformed input.
//
// Limitations:
//   - Does not support nested arrays
//   - Does not handle quoted strings containing commas
//   - Expects simple comma-separated values
func parseArray(defValue string, schema *jsonschema.Schema) any {
	// pflag represents empty slices as "[]"
	if defValue == "[]" {
		return nil
	}

	// Verify array format
	if !strings.HasPrefix(defValue, "[") || !strings.HasSuffix(defValue, "]") {
		slog.Warn("malformed array default value: must start with '[' and end with ']'", "value", defValue)
		return nil
	}

	// Remove brackets
	inner := defValue[1 : len(defValue)-1]

	// Split by comma
	parts := strings.Split(inner, ",")

	// Parse based on item type
	switch schema.Type {
	case "integer":
		return parseIntArray(parts)
	case "number":
		return parseFloatArray(parts)
	case "boolean":
		return parseBoolArray(parts)
	case "string":
		return parseStringArray(parts)
	default:
		slog.Warn("unsupported array item type for default value", "type", schema.Type, "value", defValue)
		return nil
	}
}

// parseIntArray parses a slice of strings into a slice of int64.
// Invalid elements are skipped and logged as warnings.
func parseIntArray(parts []string) []int64 {
	result := make([]int64, 0, len(parts))
	for i, p := range parts {
		trimmed := strings.TrimSpace(p)
		if val, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
			result = append(result, val)
		} else {
			slog.Warn("skipping invalid integer in array default value",
				"index", i,
				"value", p,
				"error", err)
		}
	}
	return result
}

// parseFloatArray parses a slice of strings into a slice of float64.
// Invalid elements are skipped and logged as warnings.
func parseFloatArray(parts []string) []float64 {
	result := make([]float64, 0, len(parts))
	for i, p := range parts {
		trimmed := strings.TrimSpace(p)
		if val, err := strconv.ParseFloat(trimmed, 64); err == nil {
			result = append(result, val)
		} else {
			slog.Warn("skipping invalid float in array default value",
				"index", i,
				"value", p,
				"error", err)
		}
	}
	return result
}

// parseBoolArray parses a slice of strings into a slice of bool.
// Invalid elements are skipped and logged as warnings.
func parseBoolArray(parts []string) []bool {
	result := make([]bool, 0, len(parts))
	for i, p := range parts {
		trimmed := strings.TrimSpace(p)
		if val, err := strconv.ParseBool(trimmed); err == nil {
			result = append(result, val)
		} else {
			slog.Warn("skipping invalid boolean in array default value",
				"index", i,
				"value", p,
				"error", err)
		}
	}
	return result
}

// parseStringArray parses a slice of strings, trimming whitespace from each element.
func parseStringArray(parts []string) []string {
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		result = append(result, strings.TrimSpace(p))
	}
	return result
}

func parseObject(defValue string, schema *jsonschema.Schema) any {
	if defValue == "[]" {
		return nil
	}

	// Verify array format
	if !strings.HasPrefix(defValue, "[") || !strings.HasSuffix(defValue, "]") {
		slog.Warn("malformed array default value: must start with '[' and end with ']'", "value", defValue)
		return nil
	}

	// Remove brackets
	inner := defValue[1 : len(defValue)-1]

	// Split by comma
	parts := strings.Split(inner, ",")

	// Parse based on item type
	switch schema.Type {
	case "integer":
		return parseIntObj(parts)
	case "string":
		return parseStringObj(parts)
	default:
		slog.Warn("unsupported object item type for default value", "type", schema.Type, "value", defValue)
		return nil
	}
}

func parseIntObj(parts []string) map[string]int64 {
	result := make(map[string]int64)
	for i, p := range parts {
		trimmed := strings.TrimSpace(p)
		split := strings.SplitN(trimmed, "=", 2)
		if len(split) != 2 {
			slog.Warn("malformed flag default value object", "value", p)
			continue
		}

		key := strings.TrimSpace(split[0])
		valStr := strings.TrimSpace(split[1])
		if val, err := strconv.ParseInt(valStr, 10, 64); err == nil {
			result[key] = val
		} else {
			slog.Warn("skipping invalid integer obj in array default value",
				"index", i,
				"value", p,
				"error", err)
		}
	}

	if len(result) == 0 {
		return nil
	}
	return result
}

func parseStringObj(parts []string) map[string]string {
	result := make(map[string]string)
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		split := strings.SplitN(trimmed, "=", 2)
		if len(split) != 2 {
			slog.Warn("malformed flag default value object", "value", p)
			continue
		}

		key := strings.TrimSpace(split[0])
		val := strings.TrimSpace(split[1])
		result[key] = val
	}

	if len(result) == 0 {
		return nil
	}
	return result
}
