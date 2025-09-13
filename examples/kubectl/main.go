package main

import (
	"os"

	"github.com/njayp/ophis"
	"github.com/njayp/ophis/tools"
	"k8s.io/component-base/cli"
	"k8s.io/component-base/logs"
	"k8s.io/kubectl/pkg/cmd"
	"k8s.io/kubectl/pkg/cmd/util"
)

// main taken from k8s.io/kubectl main.go
func main() {
	logs.GlogSetter(cmd.GetLogVerbosity(os.Args)) // nolint:errcheck
	command := cmd.NewDefaultKubectlCommand()

	// Add MCP server commands
	command.AddCommand(ophis.Command(&ophis.Config{
		GeneratorOptions: []tools.GeneratorOption{
			tools.WithFilters(tools.Allow([]string{
				"kubectl get",
				"kubectl describe",
				"kubectl logs",
				"kubectl top",
				"kubectl explain",
			})),
		},
	}))

	if err := cli.RunNoErrOutput(command); err != nil {
		util.CheckErr(err)
	}
}
