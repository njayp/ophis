package ophis

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

// buildCommandTree creates a command tree from a list of command names.
// The first command becomes the root, and subsequent commands are nested.
func buildCommandTree(names ...string) *cobra.Command {
	if len(names) == 0 {
		return nil
	}

	root := &cobra.Command{Use: names[0]}
	parent := root

	for _, name := range names[1:] {
		child := &cobra.Command{Use: name}
		parent.AddCommand(child)
		parent = child
	}

	return parent
}

func TestAllowCmdsContaining(t *testing.T) {
	tests := []struct {
		name         string
		allowPhrases []string
		commandNames []string
		expected     bool
	}{
		{
			name:         "matches single phrase",
			allowPhrases: []string{"get"},
			commandNames: []string{"kubectl", "get", "pods"},
			expected:     true,
		},
		{
			name:         "matches one of multiple phrases",
			allowPhrases: []string{"get", "list"},
			commandNames: []string{"helm", "list"},
			expected:     true,
		},
		{
			name:         "matches exact command name",
			allowPhrases: []string{"kubectl get"},
			commandNames: []string{"kubectl", "get", "pods"},
			expected:     true,
		},
		{
			name:         "does not match when phrase absent",
			allowPhrases: []string{"delete"},
			commandNames: []string{"kubectl", "get", "pods"},
			expected:     false,
		},
		{
			name:         "does not match any of multiple phrases",
			allowPhrases: []string{"delete", "remove"},
			commandNames: []string{"kubectl", "get", "pods"},
			expected:     false,
		},
		{
			name:         "partial match in middle of path",
			allowPhrases: []string{"admin"},
			commandNames: []string{"cli", "admin", "user"},
			expected:     true,
		},
		{
			name:         "empty phrases list rejects everything",
			allowPhrases: []string{},
			commandNames: []string{"kubectl", "get", "pods"},
			expected:     false,
		},
		{
			name:         "case sensitive matching",
			allowPhrases: []string{"Get"},
			commandNames: []string{"kubectl", "get", "pods"},
			expected:     false,
		},
		{
			name:         "matches substring in command name",
			allowPhrases: []string{"pod"},
			commandNames: []string{"kubectl", "get", "pods"},
			expected:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := buildCommandTree(tt.commandNames...)
			selector := AllowCmdsContaining(tt.allowPhrases...)
			result := selector(cmd)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExcludeCmdsContaining(t *testing.T) {
	tests := []struct {
		name           string
		excludePhrases []string
		commandNames   []string
		expected       bool
	}{
		{
			name:           "excludes matching phrase",
			excludePhrases: []string{"delete"},
			commandNames:   []string{"kubectl", "delete", "pod"},
			expected:       false,
		},
		{
			name:           "excludes when any phrase matches",
			excludePhrases: []string{"delete", "remove"},
			commandNames:   []string{"kubectl", "delete", "pod"},
			expected:       false,
		},
		{
			name:           "allows when no phrase matches",
			excludePhrases: []string{"delete"},
			commandNames:   []string{"kubectl", "get", "pods"},
			expected:       true,
		},
		{
			name:           "allows when none of multiple phrases match",
			excludePhrases: []string{"delete", "remove", "destroy"},
			commandNames:   []string{"kubectl", "get", "pods"},
			expected:       true,
		},
		{
			name:           "excludes partial match in path",
			excludePhrases: []string{"admin"},
			commandNames:   []string{"cli", "admin", "user"},
			expected:       false,
		},
		{
			name:           "empty phrases list allows everything",
			excludePhrases: []string{},
			commandNames:   []string{"kubectl", "delete", "pod"},
			expected:       true,
		},
		{
			name:           "case sensitive matching",
			excludePhrases: []string{"Delete"},
			commandNames:   []string{"kubectl", "delete", "pod"},
			expected:       true,
		},
		{
			name:           "excludes exact command name",
			excludePhrases: []string{"kubectl delete"},
			commandNames:   []string{"kubectl", "delete", "pod"},
			expected:       false,
		},
		{
			name:           "excludes substring in command name",
			excludePhrases: []string{"dele"},
			commandNames:   []string{"kubectl", "delete", "pod"},
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := buildCommandTree(tt.commandNames...)
			selector := ExcludeCmdsContaining(tt.excludePhrases...)
			result := selector(cmd)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAllowFlags(t *testing.T) {
	tests := []struct {
		name       string
		allowNames []string
		flagName   string
		expected   bool
	}{
		{
			name:       "allows single matching flag",
			allowNames: []string{"namespace"},
			flagName:   "namespace",
			expected:   true,
		},
		{
			name:       "allows one of multiple flags",
			allowNames: []string{"namespace", "output", "verbose"},
			flagName:   "output",
			expected:   true,
		},
		{
			name:       "rejects non-matching flag",
			allowNames: []string{"namespace"},
			flagName:   "kubeconfig",
			expected:   false,
		},
		{
			name:       "rejects when not in multiple allowed flags",
			allowNames: []string{"namespace", "output"},
			flagName:   "kubeconfig",
			expected:   false,
		},
		{
			name:       "empty list rejects all flags",
			allowNames: []string{},
			flagName:   "namespace",
			expected:   false,
		},
		{
			name:       "exact name match required",
			allowNames: []string{"namespace"},
			flagName:   "namespaces",
			expected:   false,
		},
		{
			name:       "case sensitive matching",
			allowNames: []string{"Namespace"},
			flagName:   "namespace",
			expected:   false,
		},
		{
			name:       "allows multiple matching flags",
			allowNames: []string{"verbose", "debug", "quiet"},
			flagName:   "debug",
			expected:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := &pflag.Flag{Name: tt.flagName}
			selector := AllowFlags(tt.allowNames...)
			result := selector(flag)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAllowCmds(t *testing.T) {
	tests := []struct {
		name         string
		allowCmds    []string
		commandNames []string
		expected     bool
	}{
		{
			name:         "matches exact command path",
			allowCmds:    []string{"kubectl get"},
			commandNames: []string{"kubectl", "get"},
			expected:     true,
		},
		{
			name:         "does not match partial path",
			allowCmds:    []string{"kubectl get"},
			commandNames: []string{"kubectl", "get", "pods"},
			expected:     false,
		},
		{
			name:         "matches one of multiple commands",
			allowCmds:    []string{"kubectl get", "helm list", "docker ps"},
			commandNames: []string{"helm", "list"},
			expected:     true,
		},
		{
			name:         "does not match if not in list",
			allowCmds:    []string{"kubectl get", "helm list"},
			commandNames: []string{"docker", "ps"},
			expected:     false,
		},
		{
			name:         "empty list rejects everything",
			allowCmds:    []string{},
			commandNames: []string{"kubectl", "get"},
			expected:     false,
		},
		{
			name:         "case sensitive matching",
			allowCmds:    []string{"kubectl Get"},
			commandNames: []string{"kubectl", "get"},
			expected:     false,
		},
		{
			name:         "requires exact match - no substring",
			allowCmds:    []string{"kubectl"},
			commandNames: []string{"kubectl", "get"},
			expected:     false,
		},
		{
			name:         "single command name",
			allowCmds:    []string{"kubectl"},
			commandNames: []string{"kubectl"},
			expected:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := buildCommandTree(tt.commandNames...)
			selector := AllowCmds(tt.allowCmds...)
			result := selector(cmd)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExcludeCmds(t *testing.T) {
	tests := []struct {
		name         string
		excludeCmds  []string
		commandNames []string
		expected     bool
	}{
		{
			name:         "excludes exact command path",
			excludeCmds:  []string{"kubectl delete"},
			commandNames: []string{"kubectl", "delete"},
			expected:     false,
		},
		{
			name:         "allows partial path not in list",
			excludeCmds:  []string{"kubectl delete"},
			commandNames: []string{"kubectl", "delete", "pod"},
			expected:     true,
		},
		{
			name:         "excludes if in list of multiple commands",
			excludeCmds:  []string{"kubectl delete", "helm uninstall", "docker rm"},
			commandNames: []string{"helm", "uninstall"},
			expected:     false,
		},
		{
			name:         "allows if not in exclusion list",
			excludeCmds:  []string{"kubectl delete", "helm uninstall"},
			commandNames: []string{"kubectl", "get"},
			expected:     true,
		},
		{
			name:         "empty list allows everything",
			excludeCmds:  []string{},
			commandNames: []string{"kubectl", "delete"},
			expected:     true,
		},
		{
			name:         "case sensitive matching",
			excludeCmds:  []string{"kubectl Delete"},
			commandNames: []string{"kubectl", "delete"},
			expected:     true,
		},
		{
			name:         "requires exact match - no substring",
			excludeCmds:  []string{"kubectl"},
			commandNames: []string{"kubectl", "delete"},
			expected:     true,
		},
		{
			name:         "single command exclusion",
			excludeCmds:  []string{"kubectl"},
			commandNames: []string{"kubectl"},
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := buildCommandTree(tt.commandNames...)
			selector := ExcludeCmds(tt.excludeCmds...)
			result := selector(cmd)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExcludeFlags(t *testing.T) {
	tests := []struct {
		name         string
		excludeNames []string
		flagName     string
		expected     bool
	}{
		{
			name:         "excludes matching flag",
			excludeNames: []string{"token"},
			flagName:     "token",
			expected:     false,
		},
		{
			name:         "excludes one of multiple flags",
			excludeNames: []string{"token", "insecure", "force"},
			flagName:     "insecure",
			expected:     false,
		},
		{
			name:         "allows non-matching flag",
			excludeNames: []string{"token"},
			flagName:     "namespace",
			expected:     true,
		},
		{
			name:         "allows when not in exclude list",
			excludeNames: []string{"token", "insecure"},
			flagName:     "namespace",
			expected:     true,
		},
		{
			name:         "empty list allows all flags",
			excludeNames: []string{},
			flagName:     "token",
			expected:     true,
		},
		{
			name:         "exact name match required",
			excludeNames: []string{"token"},
			flagName:     "tokens",
			expected:     true,
		},
		{
			name:         "case sensitive matching",
			excludeNames: []string{"Token"},
			flagName:     "token",
			expected:     true,
		},
		{
			name:         "excludes all listed flags",
			excludeNames: []string{"password", "secret", "api-key"},
			flagName:     "secret",
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := &pflag.Flag{Name: tt.flagName}
			selector := ExcludeFlags(tt.excludeNames...)
			result := selector(flag)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNoFlags(t *testing.T) {
	tests := []struct {
		name     string
		flagName string
		expected bool
	}{
		{
			name:     "rejects regular flag",
			flagName: "namespace",
			expected: false,
		},
		{
			name:     "rejects verbose flag",
			flagName: "verbose",
			expected: false,
		},
		{
			name:     "rejects output flag",
			flagName: "output",
			expected: false,
		},
		{
			name:     "rejects empty flag name",
			flagName: "",
			expected: false,
		},
		{
			name:     "rejects special character flag",
			flagName: "@special",
			expected: false,
		},
		{
			name:     "rejects long flag name",
			flagName: "very-long-flag-name-that-should-still-be-rejected",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := &pflag.Flag{Name: tt.flagName}
			result := NoFlags(flag)
			assert.Equal(t, tt.expected, result)
		})
	}
}
