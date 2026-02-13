package ophis

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func boolPtr(b bool) *bool { return &b }

func TestToolAnnotationsFromCmd(t *testing.T) {
	t.Run("no annotations returns nil", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test"}
		assert.Nil(t, toolAnnotations(cmd))
	})

	t.Run("non-MCP annotations returns nil", func(t *testing.T) {
		cmd := &cobra.Command{
			Use:         "test",
			Annotations: map[string]string{"custom": "value", "other": "data"},
		}
		assert.Nil(t, toolAnnotations(cmd))
	})

	t.Run("title", func(t *testing.T) {
		cmd := &cobra.Command{
			Use:         "test",
			Annotations: map[string]string{AnnotationTitle: "My Tool"},
		}
		ann := toolAnnotations(cmd)
		require.NotNil(t, ann)
		assert.Equal(t, "My Tool", ann.Title)
	})

	t.Run("readOnlyHint true", func(t *testing.T) {
		cmd := &cobra.Command{
			Use:         "test",
			Annotations: map[string]string{AnnotationReadOnly: "true"},
		}
		ann := toolAnnotations(cmd)
		require.NotNil(t, ann)
		assert.True(t, ann.ReadOnlyHint)
	})

	t.Run("readOnlyHint false", func(t *testing.T) {
		cmd := &cobra.Command{
			Use:         "test",
			Annotations: map[string]string{AnnotationReadOnly: "false"},
		}
		ann := toolAnnotations(cmd)
		require.NotNil(t, ann)
		assert.False(t, ann.ReadOnlyHint)
	})

	t.Run("destructiveHint true", func(t *testing.T) {
		cmd := &cobra.Command{
			Use:         "test",
			Annotations: map[string]string{AnnotationDestructive: "true"},
		}
		ann := toolAnnotations(cmd)
		require.NotNil(t, ann)
		assert.Equal(t, boolPtr(true), ann.DestructiveHint)
	})

	t.Run("destructiveHint false", func(t *testing.T) {
		cmd := &cobra.Command{
			Use:         "test",
			Annotations: map[string]string{AnnotationDestructive: "false"},
		}
		ann := toolAnnotations(cmd)
		require.NotNil(t, ann)
		assert.Equal(t, boolPtr(false), ann.DestructiveHint)
	})

	t.Run("idempotentHint", func(t *testing.T) {
		cmd := &cobra.Command{
			Use:         "test",
			Annotations: map[string]string{AnnotationIdempotent: "true"},
		}
		ann := toolAnnotations(cmd)
		require.NotNil(t, ann)
		assert.True(t, ann.IdempotentHint)
	})

	t.Run("openWorldHint true", func(t *testing.T) {
		cmd := &cobra.Command{
			Use:         "test",
			Annotations: map[string]string{AnnotationOpenWorld: "true"},
		}
		ann := toolAnnotations(cmd)
		require.NotNil(t, ann)
		assert.Equal(t, boolPtr(true), ann.OpenWorldHint)
	})

	t.Run("openWorldHint false", func(t *testing.T) {
		cmd := &cobra.Command{
			Use:         "test",
			Annotations: map[string]string{AnnotationOpenWorld: "false"},
		}
		ann := toolAnnotations(cmd)
		require.NotNil(t, ann)
		assert.Equal(t, boolPtr(false), ann.OpenWorldHint)
	})

	t.Run("all fields together", func(t *testing.T) {
		cmd := &cobra.Command{
			Use: "test",
			Annotations: map[string]string{
				AnnotationTitle:       "Delete Resource",
				AnnotationReadOnly:    "false",
				AnnotationDestructive: "true",
				AnnotationIdempotent:  "true",
				AnnotationOpenWorld:   "false",
			},
		}
		ann := toolAnnotations(cmd)
		require.NotNil(t, ann)
		assert.Equal(t, "Delete Resource", ann.Title)
		assert.False(t, ann.ReadOnlyHint)
		assert.Equal(t, boolPtr(true), ann.DestructiveHint)
		assert.True(t, ann.IdempotentHint)
		assert.Equal(t, boolPtr(false), ann.OpenWorldHint)
	})

	t.Run("invalid bool value is skipped", func(t *testing.T) {
		cmd := &cobra.Command{
			Use: "test",
			Annotations: map[string]string{
				AnnotationReadOnly:    "notabool",
				AnnotationDestructive: "invalid",
				AnnotationIdempotent:  "nope",
				AnnotationOpenWorld:   "bad",
			},
		}
		ann := toolAnnotations(cmd)
		assert.Nil(t, ann, "all invalid bools should result in nil annotations")
	})

	t.Run("mixed valid and invalid values", func(t *testing.T) {
		cmd := &cobra.Command{
			Use: "test",
			Annotations: map[string]string{
				AnnotationTitle:    "My Tool",
				AnnotationReadOnly: "notabool",
			},
		}
		ann := toolAnnotations(cmd)
		require.NotNil(t, ann, "should still return annotations for valid fields")
		assert.Equal(t, "My Tool", ann.Title)
		assert.False(t, ann.ReadOnlyHint, "invalid readOnlyHint should remain zero value")
	})

	t.Run("strconv.ParseBool variants", func(t *testing.T) {
		for _, v := range []string{"1", "t", "TRUE", "True"} {
			cmd := &cobra.Command{
				Use:         "test",
				Annotations: map[string]string{AnnotationReadOnly: v},
			}
			ann := toolAnnotations(cmd)
			require.NotNil(t, ann, "value %q should parse as true", v)
			assert.True(t, ann.ReadOnlyHint, "value %q should parse as true", v)
		}
		for _, v := range []string{"0", "f", "FALSE", "False"} {
			cmd := &cobra.Command{
				Use:         "test",
				Annotations: map[string]string{AnnotationDestructive: v},
			}
			ann := toolAnnotations(cmd)
			require.NotNil(t, ann, "value %q should parse as false", v)
			assert.Equal(t, boolPtr(false), ann.DestructiveHint, "value %q should parse as false", v)
		}
	})
}

func TestCreateToolFromCmd_Annotations(t *testing.T) {
	t.Run("annotations propagated to tool", func(t *testing.T) {
		cmd := &cobra.Command{
			Use:   "delete",
			Short: "Delete a resource",
			Run:   func(_ *cobra.Command, _ []string) {},
			Annotations: map[string]string{
				AnnotationTitle:       "Delete Resource",
				AnnotationDestructive: "true",
				AnnotationOpenWorld:   "false",
			},
		}
		root := &cobra.Command{Use: "app"}
		root.AddCommand(cmd)

		tool := Selector{}.createToolFromCmd(cmd, "app")
		require.NotNil(t, tool.Annotations)
		assert.Equal(t, "Delete Resource", tool.Annotations.Title)
		assert.Equal(t, boolPtr(true), tool.Annotations.DestructiveHint)
		assert.Equal(t, boolPtr(false), tool.Annotations.OpenWorldHint)
	})

	t.Run("no annotations results in nil", func(t *testing.T) {
		cmd := &cobra.Command{
			Use:   "list",
			Short: "List resources",
			Run:   func(_ *cobra.Command, _ []string) {},
		}
		root := &cobra.Command{Use: "app"}
		root.AddCommand(cmd)

		tool := Selector{}.createToolFromCmd(cmd, "app")
		assert.Nil(t, tool.Annotations)
	})
}
