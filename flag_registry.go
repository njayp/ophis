package ophis

// FlagMetadata tracks metadata about a flag for proper serialization
type FlagMetadata struct {
	// HasJSONSchema indicates if this flag uses a JSON schema annotation
	// and should be marshaled as JSON instead of key-value pairs
	HasJSONSchema bool
}

// FlagMetadataByFlagName contains metadata for all flags by flag name for a single tool
type FlagMetadataByFlagName map[string]FlagMetadata

// FlagRegistry stores Metadata about all tool flags
// It is populated once during server initialization and then remains read-only
type FlagRegistry struct {
	// tools maps tool name -> flag name -> metadata
	tools map[string]FlagMetadataByFlagName
}

// NewFlagRegistry creates a new empty flag registry
func NewFlagRegistry() *FlagRegistry {
	return &FlagRegistry{
		tools: make(map[string]FlagMetadataByFlagName),
	}
}

// Register stores flag metadata for a tool
func (r *FlagRegistry) Register(toolName string, flags FlagMetadataByFlagName) {
	r.tools[toolName] = flags
}

// HasJSONSchema checks if a specific flag has a JSON schema annotation
// Returns false if tool or flag is not found
func (r *FlagRegistry) HasJSONSchema(toolName string, flagName string) bool {
	if flags, ok := r.tools[toolName]; ok {
		if meta, ok := flags[flagName]; ok {
			return meta.HasJSONSchema
		}
	}
	return false
}

// globalFlagRegistry is populated during server initialization
var globalFlagRegistry = NewFlagRegistry()
