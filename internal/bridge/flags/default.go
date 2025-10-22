package flags

import (
	"encoding/json"
	"log/slog"
	"strconv"
	"strings"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/spf13/pflag"
)

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
