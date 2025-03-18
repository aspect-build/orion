load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")
load("//cli/core/gazelle:deps.bzl", "fetch_deps")

fetch_deps()

http_archive(
    name = "io_bazel_rules_go",
    sha256 = "90fe8fb402dee957a375f3eb8511455bd738c7ed562695f4dd117ac7d2d833b1",
    urls = [
        "https://mirror.bazel.build/github.com/bazel-contrib/rules_go/releases/download/v0.52.0/rules_go-v0.52.0.zip",
        "https://github.com/bazel-contrib/rules_go/releases/download/v0.52.0/rules_go-v0.52.0.zip",
    ],
)


load("@io_bazel_rules_go//go:deps.bzl", "go_register_nogo", "go_register_toolchains", "go_rules_dependencies")
load("//:deps.bzl", "go_dependencies")

# gazelle:repository_macro deps.bzl%go_dependencies
go_dependencies()

go_rules_dependencies()

go_register_toolchains(
    version = "1.24.1",
)

go_register_nogo(
    includes = ["@//:__subpackages__"],
    nogo = "@//bazel/go:nogo",
)


load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies")
gazelle_dependencies()



http_archive(
    name = "rules_python",
    sha256 = "778aaeab3e6cfd56d681c89f5c10d7ad6bf8d2f1a72de9de55b23081b2d31618",
    # NB: version matches go.mod where we fetch the rules_python/gazelle Go package.
    # Run: `go get github.com/bazelbuild/rules_python/gazelle@x.x.x`
    strip_prefix = "rules_python-0.34.0",
    url = "https://github.com/bazelbuild/rules_python/releases/download/0.34.0/rules_python-0.34.0.tar.gz",
    patch_args = ["-p1"],
    patches = ["//third_party:github.com/bazelbuild/rules_python/bzlmod.patch"],
)
# Needed to compile the gazelle extension
load("@rules_python//gazelle:deps.bzl", "python_stdlib_list_deps")
python_stdlib_list_deps()
