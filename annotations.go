package ophis

import (
	"log/slog"
	"strconv"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

// Cobra command annotation keys for MCP tool annotations.
// Set these in cmd.Annotations to populate mcp.ToolAnnotations on the generated tool.
//
// Boolean values are parsed with strconv.ParseBool (accepts "1", "t", "true", "0", "f", "false", etc.).
//
// Example:
//
//	cmd.Annotations = map[string]string{
//	    ophis.AnnotationReadOnly: "true",
//	    ophis.AnnotationTitle:    "List files",
//	}
const (
	// AnnotationTitle sets the human-readable title for the tool.
	AnnotationTitle = "title"

	// AnnotationReadOnly hints that the tool does not modify its environment.
	AnnotationReadOnly = "readOnlyHint"

	// AnnotationDestructive hints that the tool may perform destructive updates.
	// Only meaningful when ReadOnlyHint is false.
	AnnotationDestructive = "destructiveHint"

	// AnnotationIdempotent hints that calling the tool repeatedly with the same
	// arguments has no additional effect. Only meaningful when ReadOnlyHint is false.
	AnnotationIdempotent = "idempotentHint"

	// AnnotationOpenWorld hints that the tool may interact with external entities
	// outside its closed domain.
	AnnotationOpenWorld = "openWorldHint"
)

// toolAnnotations reads MCP annotation keys from cmd.Annotations and
// returns a populated *mcp.ToolAnnotations, or nil if no MCP annotations are found.
func toolAnnotations(cmd *cobra.Command) *mcp.ToolAnnotations {
	if len(cmd.Annotations) == 0 {
		return nil
	}

	var annotations mcp.ToolAnnotations
	found := false

	if v, ok := cmd.Annotations[AnnotationTitle]; ok {
		annotations.Title = v
		found = true
	}

	if v, ok := cmd.Annotations[AnnotationReadOnly]; ok {
		b, err := strconv.ParseBool(v)
		if err != nil {
			slog.Warn("invalid bool value for annotation, skipping", "key", AnnotationReadOnly, "value", v)
		} else {
			annotations.ReadOnlyHint = b
			found = true
		}
	}

	if v, ok := cmd.Annotations[AnnotationDestructive]; ok {
		b, err := strconv.ParseBool(v)
		if err != nil {
			slog.Warn("invalid bool value for annotation, skipping", "key", AnnotationDestructive, "value", v)
		} else {
			annotations.DestructiveHint = &b
			found = true
		}
	}

	if v, ok := cmd.Annotations[AnnotationIdempotent]; ok {
		b, err := strconv.ParseBool(v)
		if err != nil {
			slog.Warn("invalid bool value for annotation, skipping", "key", AnnotationIdempotent, "value", v)
		} else {
			annotations.IdempotentHint = b
			found = true
		}
	}

	if v, ok := cmd.Annotations[AnnotationOpenWorld]; ok {
		b, err := strconv.ParseBool(v)
		if err != nil {
			slog.Warn("invalid bool value for annotation, skipping", "key", AnnotationOpenWorld, "value", v)
		} else {
			annotations.OpenWorldHint = &b
			found = true
		}
	}

	if !found {
		return nil
	}

	return &annotations
}
