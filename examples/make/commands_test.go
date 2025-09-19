package main

import (
	"testing"

	"github.com/njayp/ophis/examples"
)

func TestTools(t *testing.T) {
	cmd := createMakeCommands()

	examples.TestTools(t, cmd, []string{
		"make_lint",
		"make_test",
	})
}
