package helm

import (
	"io"

	"github.com/spf13/cobra"
	helmcmd "helm.sh/helm/v4/pkg/cmd"
)

func NewHelmCommand(output io.Writer) *cobra.Command {
	cmd, err := helmcmd.NewRootCmd(output, nil)
	if err != nil {
		panic(err)
	}
	return cmd
}
