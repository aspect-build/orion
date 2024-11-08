workspace(name = "orion")

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

# Ensure this version always matches the go.mod version.
http_archive(
    name = "io_bazel_rules_go",
    sha256 = "f4a9314518ca6acfa16cc4ab43b0b8ce1e4ea64b81c38d8a3772883f153346b8",
    urls = ["https://github.com/bazelbuild/rules_go/releases/download/v0.50.1/rules_go-v0.50.1.zip"],
)

# Ensure this version always matches the go.mod version.
http_archive(
    name = "bazel_gazelle",
    sha256 = "795b1d0b22630aa218e48e405de9b63f9292a5850b1a06552fad1d9acecadfc7",
    strip_prefix = "bazel-gazelle-da0ad1847d05e62366a83bb51d965da413687858",
    urls = ["https://github.com/bazelbuild/bazel-gazelle/archive/da0ad1847d05e62366a83bb51d965da413687858.tar.gz"],
    patch_args = ["-p1"],
    patches = [
        "//cli/core:patches/bazelbuild_bazel-gazelle_aspect-cli.patch",
        "//cli/core:patches/bazelbuild_bazel-gazelle_aspect-walk-subdir.patch",
        "//cli/core:patches/bazelbuild_bazel-gazelle_aspect-gitignore.patch",
        "//cli/core:patches/bazelbuild_bazel-gazelle_aspect-fs-direntry.patch",
    ],
)

load("//cli/core/gazelle/common/treesitter/grammars:grammars.bzl", "fetch_grammars")

fetch_grammars()

load("@io_bazel_rules_go//go:deps.bzl", "go_register_nogo", "go_register_toolchains", "go_rules_dependencies")
load("//:deps.bzl", "go_dependencies")

# gazelle:repository_macro deps.bzl%go_dependencies
go_dependencies()

go_rules_dependencies()

go_register_toolchains(
    version = "1.23.0",
)

go_register_nogo(
    includes = ["@//:__subpackages__"],
    nogo = "@//bazel/go:nogo",
)

load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies")
gazelle_dependencies()
