package main

import (
	"testing"

	"github.com/njayp/ophis/test"
)

func TestTools(t *testing.T) {
	test.Tools(t, makeCmd(), "make")
}
