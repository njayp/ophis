package main

import (
	"testing"

	"github.com/njayp/ophis/examples"
)

func TestTools(t *testing.T) {
	cmd := rootCmd()

	examples.TestTools(t, cmd, []string{
		"argocd_app_get",
		"argocd_app_list",
		"argocd_app_diff",
		"argocd_app_manifests",
		"argocd_app_history",
		"argocd_app_resources",
		"argocd_app_logs",
	})
}
