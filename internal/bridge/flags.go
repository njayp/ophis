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
		// Handle array types (slices)
		// pflag represents empty slices as "[]"
		if defValue == "[]" {
			return
		}
		// pflag represents arrays as "[item1,item2,item3]"
		// We need to manually parse this into an actual JSON array
		// --- Ewwww Gross ---
		if strings.HasPrefix(defValue, "[") && strings.HasSuffix(defValue, "]") {
			// Remove the brackets
			inner := defValue[1 : len(defValue)-1]
			if inner == "" {
				return // Empty array
			}
			// Split by comma
			parts := strings.Split(inner, ",")
			// Determine the array item type from the schema
			if flagSchema.Items != nil {
				switch flagSchema.Items.Type {
				case "integer":
					// Parse as integer array
					intArr := make([]int64, 0, len(parts))
					for _, p := range parts {
						if val, err := strconv.ParseInt(strings.TrimSpace(p), 10, 64); err == nil {
							intArr = append(intArr, val)
						}
					}
					setDefault(intArr)
				case "number":
					// Parse as float array
					floatArr := make([]float64, 0, len(parts))
					for _, p := range parts {
						if val, err := strconv.ParseFloat(strings.TrimSpace(p), 64); err == nil {
							floatArr = append(floatArr, val)
						}
					}
					setDefault(floatArr)
				case "boolean":
					// Parse as boolean array
					boolArr := make([]bool, 0, len(parts))
					for _, p := range parts {
						if val, err := strconv.ParseBool(strings.TrimSpace(p)); err == nil {
							boolArr = append(boolArr, val)
						}
					}
					setDefault(boolArr)
				case "string":
					// String array - trim whitespace from each element
					strArr := make([]string, 0, len(parts))
					for _, p := range parts {
						strArr = append(strArr, strings.TrimSpace(p))
					}
					setDefault(strArr)
				}

				// there are no array of objects in pflag, so we don't handle that case
			}
		}
	}
}
