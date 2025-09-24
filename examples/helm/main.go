package main

import (
	"log/slog"
	"os"

	// Import to initialize client auth plugins.
	"github.com/njayp/ophis"
	"github.com/spf13/cobra"
	helmcmd "helm.sh/helm/v4/pkg/cmd"
	"helm.sh/helm/v4/pkg/kube"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func rootCmd() *cobra.Command {
	cmd, err := helmcmd.NewRootCmd(os.Stdout, os.Args[1:], helmcmd.SetupLogging)
	if err != nil {
		slog.Warn("command failed", slog.Any("error", err))
		os.Exit(1)
	}

	// add mcp server commands
	cmd.AddCommand(ophis.Command(&ophis.Config{
		Selectors: []ophis.Selector{
			{
				CmdSelect: ophis.AllowCmd(
					"helm list",
					"helm status",
					"helm get",
					"helm history",
					"helm show",
					"helm search",
				),
			},
		},
	}))

	return cmd
}

// main taken from https://github.com/helm/helm/blob/main/cmd/helm/helm.go
func main() {
	kube.ManagedFieldsManager = "helm"

	cmd := rootCmd()
	if err := cmd.Execute(); err != nil {
		if cerr, ok := err.(helmcmd.CommandError); ok {
			os.Exit(cerr.ExitCode)
		}
		os.Exit(1)
	}
}
