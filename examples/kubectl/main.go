package main

import (
	"os"

	"github.com/njayp/ophis"
	"github.com/spf13/cobra"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/component-base/cli"
	"k8s.io/component-base/logs"
	"k8s.io/kubectl/pkg/cmd"
	"k8s.io/kubectl/pkg/cmd/util"
)

func rootCmd() *cobra.Command {
	command := cmd.NewDefaultKubectlCommand()

	// Add MCP server commands
	command.AddCommand(ophis.Command(&ophis.Config{
		Filters: []ophis.Filter{ophis.Allow([]string{
			"kubectl get",
			"kubectl describe",
			"kubectl logs",
			"kubectl top pod",
			"kubectl top node",
			"kubectl explain",
		})},
	}))

	return command
}

// main taken from https://github.com/kubernetes/kubernetes/blob/master/cmd/kubectl/kubectl.go
func main() {
	logs.GlogSetter(cmd.GetLogVerbosity(os.Args)) // nolint:errcheck

	command := rootCmd()
	if err := cli.RunNoErrOutput(command); err != nil {
		util.CheckErr(err)
	}
}
