package ophis

import (
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func TestHiddenFlagFilter(t *testing.T) {
	filter := hiddenFlagFilter()

	tests := []struct {
		name     string
		flag     *pflag.Flag
		expected bool
	}{
		{
			name: "visible flag passes",
			flag: &pflag.Flag{
				Name:   "visible",
				Hidden: false,
			},
			expected: true,
		},
		{
			name: "hidden flag is excluded",
			flag: &pflag.Flag{
				Name:   "hidden",
				Hidden: true,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filter(tt.flag)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDepreciatedFlagFilter(t *testing.T) {
	filter := depreciatedFlagFilter()

	tests := []struct {
		name     string
		flag     *pflag.Flag
		expected bool
	}{
		{
			name: "non-deprecated flag passes",
			flag: &pflag.Flag{
				Name:       "current",
				Deprecated: "",
			},
			expected: true,
		},
		{
			name: "deprecated flag is excluded",
			flag: &pflag.Flag{
				Name:       "old",
				Deprecated: "use --new instead",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filter(tt.flag)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExcludeFlagFilter(t *testing.T) {
	tests := []struct {
		name        string
		excludeList []string
		flag        *pflag.Flag
		expected    bool
	}{
		{
			name:        "exact match excludes",
			excludeList: []string{"debug"},
			flag: &pflag.Flag{
				Name: "debug",
			},
			expected: false,
		},
		{
			name:        "partial match excludes",
			excludeList: []string{"debug"},
			flag: &pflag.Flag{
				Name: "debug-level",
			},
			expected: false,
		},
		{
			name:        "no match passes",
			excludeList: []string{"debug", "trace"},
			flag: &pflag.Flag{
				Name: "verbose",
			},
			expected: true,
		},
		{
			name:        "empty exclude list passes all",
			excludeList: []string{},
			flag: &pflag.Flag{
				Name: "anything",
			},
			expected: true,
		},
		{
			name:        "multiple excludes work",
			excludeList: []string{"admin", "danger", "internal"},
			flag: &pflag.Flag{
				Name: "danger-zone",
			},
			expected: false,
		},
		{
			name:        "case sensitive matching",
			excludeList: []string{"Debug"},
			flag: &pflag.Flag{
				Name: "debug",
			},
			expected: true, // Should pass because case doesn't match
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := ExcludeFlagFilter(tt.excludeList...)
			result := filter(tt.flag)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAllowFlagFilter(t *testing.T) {
	tests := []struct {
		name      string
		allowList []string
		flag      *pflag.Flag
		expected  bool
	}{
		{
			name:      "exact match passes",
			allowList: []string{"verbose"},
			flag: &pflag.Flag{
				Name: "verbose",
			},
			expected: true,
		},
		{
			name:      "partial match passes",
			allowList: []string{"log"},
			flag: &pflag.Flag{
				Name: "log-level",
			},
			expected: true,
		},
		{
			name:      "no match excludes",
			allowList: []string{"verbose", "debug"},
			flag: &pflag.Flag{
				Name: "quiet",
			},
			expected: false,
		},
		{
			name:      "empty allow list excludes all",
			allowList: []string{},
			flag: &pflag.Flag{
				Name: "anything",
			},
			expected: false,
		},
		{
			name:      "multiple allows work",
			allowList: []string{"output", "format", "verbose"},
			flag: &pflag.Flag{
				Name: "output-format",
			},
			expected: true, // Matches "output"
		},
		{
			name:      "case sensitive matching",
			allowList: []string{"Verbose"},
			flag: &pflag.Flag{
				Name: "verbose",
			},
			expected: false, // Should fail because case doesn't match
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := AllowFlagFilter(tt.allowList...)
			result := filter(tt.flag)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCombinedFilters(t *testing.T) {
	// Test that filters can be combined effectively
	t.Run("exclude and allow filters combined", func(t *testing.T) {
		// Create a flag that should pass allow filter but fail exclude filter
		flag := &pflag.Flag{
			Name: "debug-output",
		}

		allowFilter := AllowFlagFilter("output")    // Should pass
		excludeFilter := ExcludeFlagFilter("debug") // Should fail

		// In the actual implementation, both filters would be applied
		// and the flag would need to pass ALL filters
		assert.True(t, allowFilter(flag), "Should pass allow filter")
		assert.False(t, excludeFilter(flag), "Should fail exclude filter")
	})

	t.Run("hidden flag with allow filter", func(t *testing.T) {
		hiddenFlag := &pflag.Flag{
			Name:   "allowed-but-hidden",
			Hidden: true,
		}

		allowFilter := AllowFlagFilter("allowed")
		hiddenFilter := hiddenFlagFilter()

		assert.True(t, allowFilter(hiddenFlag), "Should pass allow filter")
		assert.False(t, hiddenFilter(hiddenFlag), "Should fail hidden filter")
	})

	t.Run("deprecated flag with allow filter", func(t *testing.T) {
		deprecatedFlag := &pflag.Flag{
			Name:       "allowed-but-deprecated",
			Deprecated: "use something else",
		}

		allowFilter := AllowFlagFilter("allowed")
		deprecatedFilter := depreciatedFlagFilter()

		assert.True(t, allowFilter(deprecatedFlag), "Should pass allow filter")
		assert.False(t, deprecatedFilter(deprecatedFlag), "Should fail deprecated filter")
	})
}
