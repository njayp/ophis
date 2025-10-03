package main

import (
	"testing"

	"github.com/njayp/ophis/examples"
)

func Test_main(t *testing.T) {
	examples.TestTools(t, rootCmd(), []string{
		"argocd_app_get",
		"argocd_app_list",
		"argocd_app_diff",
		"argocd_app_manifests",
		"argocd_app_history",
		"argocd_app_resources",
		"argocd_app_logs",
		"argocd_app_sync",
		"argocd_app_wait",
		"argocd_app_rollback",
		"argocd_cluster_list",
		"argocd_proj_list",
		"argocd_repo_list",
	})
}
