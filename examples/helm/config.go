package main

import "github.com/njayp/ophis"

func config() *ophis.Config {
	return &ophis.Config{
		Selectors: selectors(),
	}
}

func selectors() []ophis.Selector {
	return []ophis.Selector{
		{
			CmdSelector: ophis.AllowCmds(
				"helm get hooks",
				"helm get manifest",
				"helm get notes",
				"helm get values",
				"helm history",
				"helm list",
				"helm search hub",
				"helm search repo",
				"helm show chart",
				"helm show crds",
				"helm show readme",
				"helm show values",
				"helm status",
				"helm repo list",
				"helm template",
				"helm dependency list",
				"helm lint",
			),

			InheritedFlagSelector: ophis.AllowFlags("namespace"),
		},
	}
}
