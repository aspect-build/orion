package gazelle

import plugin "github.com/aspect-build/silo/cli/pro/gazelle/host/plugin"

var builtinKinds = []plugin.RuleKind{
	// @aspect_bazel_lib
	plugin.RuleKind{
		Name: "copy_to_bin",
		From: "@aspect_bazel_lib//lib:copy_to_bin.bzl",
	},
}
