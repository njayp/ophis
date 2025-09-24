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
			CmdSelector: ophis.AllowCmd("helm get"),
			FlagSelector: ophis.AllowFlag(
				"namespace",
				"revision",
				"all",
				"output",
			),
		},
		{
			CmdSelector: ophis.AllowCmd("helm history"),
			FlagSelector: ophis.AllowFlag(
				"namespace",
				"max",
				"output",
			),
		},
		{
			CmdSelector: ophis.AllowCmd("helm list"),
			FlagSelector: ophis.AllowFlag(
				"all-namespaces",
				"namespace",
				"filter",
				"output",
				"deployed",
				"failed",
				"pending",
				"uninstalled",
				"superseded",
				"uninstalling",
				"date",
				"reverse",
				"selector",
				"short",
			),
		},
		{
			CmdSelector: ophis.AllowCmd("helm search hub"),
			FlagSelector: ophis.AllowFlag(
				"max-col-width",
				"output",
				"list-repo-url",
			),
		},
		{
			CmdSelector: ophis.AllowCmd("helm search repo"),
			FlagSelector: ophis.AllowFlag(
				"regexp",
				"version",
				"versions",
				"output",
				"devel",
			),
		},
		{
			CmdSelector: ophis.AllowCmd("helm show"),
			FlagSelector: ophis.AllowFlag(
				"version",
				"devel",
				"verify",
				"insecure-skip-tls-verify",
				"ca-file",
				"cert-file",
				"key-file",
				"pass-credentials",
				"repo",
				"jsonpath",
			),
		},
		{
			CmdSelector: ophis.AllowCmd("helm status"),
			FlagSelector: ophis.AllowFlag(
				"namespace",
				"revision",
				"output",
				"show-desc",
				"show-resources",
			),
		},
	}
}
