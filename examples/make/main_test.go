package main

import (
	"testing"

	"github.com/njayp/ophis/examples"
)

func TestTools(t *testing.T) {
	cmd := makeCmd()

	examples.TestTools(t, cmd, []string{
		"make",
	})
}
