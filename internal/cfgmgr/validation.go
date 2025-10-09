package cfgmgr

import (
	"fmt"
	"strings"
)

// ValidLogLevels are the accepted log level values
var ValidLogLevels = []string{"debug", "info", "warn", "error"}

// ValidateLogLevel checks if the provided log level is valid
func ValidateLogLevel(level string) error {
	if level == "" {
		return nil // empty is allowed (means not set)
	}

	normalizedLevel := strings.ToLower(level)
	for _, valid := range ValidLogLevels {
		if normalizedLevel == valid {
			return nil
		}
	}

	return fmt.Errorf("invalid log level %q: must be one of %v", level, ValidLogLevels)
}

// ValidateServerName checks if a server name is valid
func ValidateServerName(name string) error {
	if name == "" {
		return fmt.Errorf("server name cannot be empty")
	}

	// Check for invalid characters (basic validation)
	if strings.ContainsAny(name, "/\\:*?\"<>|") {
		return fmt.Errorf("server name contains invalid characters: %s", name)
	}

	return nil
}
