package ophis

import (
	"os"
	"testing"
)

func TestConfig(t *testing.T) {
	cache, _ := os.UserCacheDir()
	t.Error(cache)
}
