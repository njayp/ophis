package ophis

import (
	"fmt"

	"github.com/spf13/pflag"
)

func descriptionFromFlag(flag *pflag.Flag) string {
	description := flag.Usage
	if description == "" {
		description = fmt.Sprintf("Flag: %s", flag.Name)
	}

	return description
}

func flagToolOption(flag *pflag.Flag) map[string]string {
	return map[string]string{
		"type":        flag.Value.Type(),
		"description": descriptionFromFlag(flag),
	}
}
