package main

import "github.com/njayp/ophis"

func config() *ophis.Config {
	return &ophis.Config{
		Selectors: selectors(),
	}
}

func allowNamespace() ophis.FlagSelector {
	return ophis.AllowFlag("namespace")
}

func selectors() []ophis.Selector {
	return []ophis.Selector{
		{
			// helm get [all, hooks, manifest, metadata, notes, values]
			CmdSelector:           ophis.AllowCmd("helm get"),
			InheritedFlagSelector: allowNamespace(),
		},
		{
			CmdSelector:           ophis.AllowCmd("helm history"),
			InheritedFlagSelector: allowNamespace(),
		},
		{
			CmdSelector:           ophis.AllowCmd("helm list"),
			InheritedFlagSelector: allowNamespace(),
		},
		{
			CmdSelector:           ophis.AllowCmd("helm search hub"),
			InheritedFlagSelector: allowNamespace(),
		},
		{
			CmdSelector:           ophis.AllowCmd("helm search repo"),
			InheritedFlagSelector: allowNamespace(),
		},
		{
			// helm show [all, chart, crds, readme, values]
			CmdSelector:           ophis.AllowCmd("helm show"),
			InheritedFlagSelector: allowNamespace(),
		},
		{
			CmdSelector:           ophis.AllowCmd("helm status"),
			InheritedFlagSelector: allowNamespace(),
		},
	}
}
