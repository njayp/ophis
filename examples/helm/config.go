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
			CmdSelector: ophis.AllowCmd(
				// helm get [all, hooks, manifest, metadata, notes, values]
				"helm get",
				"helm history",
				"helm list",
				"helm search repo",
				"helm search hub",
				// helm show [all, chart, crds, readme, values]
				"helm show",
				"helm status",
			),
			InheritedFlagSelector: ophis.AllowFlag("namespace"),
		},
	}
}
