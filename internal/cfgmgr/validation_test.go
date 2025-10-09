package cfgmgr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateLogLevel(t *testing.T) {
	tests := []struct {
		name      string
		level     string
		shouldErr bool
	}{
		{"empty is valid", "", false},
		{"debug is valid", "debug", false},
		{"info is valid", "info", false},
		{"warn is valid", "warn", false},
		{"error is valid", "error", false},
		{"uppercase debug is valid", "DEBUG", false},
		{"mixed case info is valid", "Info", false},
		{"invalid level", "invalid", true},
		{"trace is invalid", "trace", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateLogLevel(tt.level)
			if tt.shouldErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateServerName(t *testing.T) {
	tests := []struct {
		name       string
		serverName string
		shouldErr  bool
	}{
		{"valid simple name", "my-server", false},
		{"valid with numbers", "server123", false},
		{"valid with underscores", "my_server", false},
		{"valid with dots", "my.server", false},
		{"empty name", "", true},
		{"with forward slash", "my/server", true},
		{"with backslash", "my\\server", true},
		{"with colon", "my:server", true},
		{"with asterisk", "my*server", true},
		{"with question mark", "my?server", true},
		{"with quotes", "my\"server", true},
		{"with angle brackets", "my<server>", true},
		{"with pipe", "my|server", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateServerName(tt.serverName)
			if tt.shouldErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
