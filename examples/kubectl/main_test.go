package main

import (
	"testing"

	"github.com/njayp/ophis/test"
)

func TestTools(t *testing.T) {
	cmd := rootCmd()

	test.Tools(t, cmd,
		"kubectl_get",
		"kubectl_describe",
		"kubectl_logs",
		"kubectl_explain",
		"kubectl_api-resources",
	)
}
