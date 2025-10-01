package main

import (
	"errors"
	"os"
	"os/exec"

	cli "github.com/argoproj/argo-cd/v3/cmd/argocd/commands"
	"github.com/argoproj/argo-cd/v3/util/log"
	"github.com/njayp/ophis"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

func init() {
	// Make sure klog uses the configured log level and format.
	klog.SetLogger(log.NewLogrusLogger(log.NewWithCurrentConfig()))
}

func rootCmd() *cobra.Command {
	command := cli.NewCommand()
	command.AddCommand(ophis.Command(&ophis.Config{
		Selectors: []ophis.Selector{
			{
				CmdSelector: ophis.AllowCmd(
					"argocd app get",
					"argocd app list",
					"argocd app diff",
					"argocd app manifests",
					"argocd app history",
					"argocd app resources",
					"argocd app logs",
				),

				InheritedFlagSelector: ophis.NoFlags,
			},
		},
	}))
	return command
}

// main from https://github.com/argoproj/argo-cd/blob/master/cmd/main.go
func main() {
	isArgocdCLI := true
	command := rootCmd()
	command.SilenceErrors = true
	command.SilenceUsage = true

	err := command.Execute()
	// if the err is non-nil, try to look for various scenarios
	// such as if the error is from the execution of a normal argocd command,
	// unknown command error or any other.
	if err != nil {
		pluginHandler := cli.NewDefaultPluginHandler([]string{"argocd"})
		pluginErr := pluginHandler.HandleCommandExecutionError(err, isArgocdCLI, os.Args)
		if pluginErr != nil {
			var exitErr *exec.ExitError
			if errors.As(pluginErr, &exitErr) {
				// Return the actual plugin exit code
				os.Exit(exitErr.ExitCode())
			}
			// Fallback to exit code 1 if the error isn't an exec.ExitError
			os.Exit(1)
		}
	}
}
