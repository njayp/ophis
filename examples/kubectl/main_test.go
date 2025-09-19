package main

import (
	"testing"

	"github.com/njayp/ophis/examples"
)

func TestTools(t *testing.T) {
	cmd := rootCmd()

	examples.TestTools(t, cmd, []string{
		"kubectl_get",
		"kubectl_describe",
		"kubectl_logs",
		"kubectl_top_pod",
		"kubectl_top_node",
		"kubectl_explain",
	})
}
