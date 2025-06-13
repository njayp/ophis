package ophis

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func descendCmdTree(cmd *cobra.Command, cmdPath string) (*cobra.Command, error) {
	fields := strings.Fields(cmdPath)

	// flags must be set on relevant command
	if len(fields) > 1 {
		// move to subCommand
		for _, field := range fields {
			for _, subCmd := range cmd.Commands() {
				if field == subCmd.Name() {
					cmd = subCmd
					break
				}
			}
		}
	}

	// verify cmd is set
	newPath := cmd.CommandPath()
	if newPath != cmdPath {
		return nil, fmt.Errorf("command path not recognized: %s", cmdPath)
	}
	return cmd, nil
}
