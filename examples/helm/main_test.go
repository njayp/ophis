package main

import (
	"testing"

	"github.com/njayp/ophis/examples"
)

func TestTools(t *testing.T) {
	cmd := rootCmd()

	examples.TestTools(t, cmd, []string{
		"helm_get_hooks",
		"helm_get_manifest",
		"helm_get_notes",
		"helm_get_values",
		"helm_history",
		"helm_list",
		"helm_search_hub",
		"helm_search_repo",
		"helm_show_chart",
		"helm_show_crds",
		"helm_show_readme",
		"helm_show_values",
		"helm_status",
		"helm_repo_list",
		"helm_template",
		"helm_dependency_list",
		"helm_lint",
	})
}
