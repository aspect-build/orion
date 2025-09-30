load("@bazelrc-preset.bzl", "bazelrc_preset")
load("@buildifier_prebuilt//:rules.bzl", "buildifier")

bazelrc_preset(name = "preset")

buildifier(
    name = "buildifier",
    exclude_patterns = ["./.git/*"],
    lint_mode = "fix",
    mode = "fix",
    visibility = ["//visibility:public"],
)

buildifier(
    name = "buildifier.check",
    exclude_patterns = ["./.git/*"],
    lint_mode = "warn",
    mode = "diff",
    visibility = ["//visibility:public"],
)
