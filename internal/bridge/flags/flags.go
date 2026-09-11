package flags

import (
	"fmt"
	"log/slog"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// FlagAnnotationJSONSchema identifies a flag's JSON schema override.
const FlagAnnotationJSONSchema = "jsonschema"

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

// AddFlagToSchema adds a single flag to the schema properties.
func AddFlagToSchema(schema *jsonschema.Schema, flag *pflag.Flag) {
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
	if values, ok := flag.Annotations[FlagAnnotationJSONSchema]; ok {
		if len(values) == 0 {
			slog.Warn(fmt.Sprintf("No value for jsonschema annotation for flag %s, using its native type", flag.Name))
		} else {
			var annotatedSchema jsonschema.Schema
			if err := annotatedSchema.UnmarshalJSON([]byte(values[0])); err != nil {
				slog.Error(fmt.Sprintf("Error when decoding JSON schema for flag %s (%v), using its native type", flag.Name, err))
			} else {
				if annotatedSchema.Description == "" {
					annotatedSchema.Description = flag.Usage
				}
				setDefaultFromFlag(&annotatedSchema, flag)
				validateDefault(&annotatedSchema, flag)
				embeddedSchema := &jsonschema.Schema{Properties: map[string]*jsonschema.Schema{
					"flags": {Properties: map[string]*jsonschema.Schema{flag.Name: &annotatedSchema}},
				}}
				if _, err := embeddedSchema.Resolve(&jsonschema.ResolveOptions{ValidateDefaults: true}); err != nil {
					slog.Error(fmt.Sprintf("Annotated JSON schema for flag %s is unusable (%v), using its native type", flag.Name, err))
				} else {
					schema.Properties[flag.Name] = &annotatedSchema
					return
				}
			}
		}
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
	// pflag validates IP values; regexes reject valid compressed IPv6 addresses.
	case "ip":
		flagSchema.Type = "string"
		flagSchema.Description += " (format: IPv4 or IPv6 address)"
	case "ipMask":
		flagSchema.Type = "string"
		flagSchema.Description += " (format: IP mask, e.g., '255.255.255.0')"
	case "ipNet":
		flagSchema.Type = "string"
		flagSchema.Description += " (format: CIDR notation, e.g., '192.168.1.0/24')"
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
	validateDefault(flagSchema, flag)
	schema.Properties[flag.Name] = flagSchema
}
