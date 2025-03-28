echo "Generate base Orion repository"

# Gitignore bazel stuff
cat >.gitignore <<-EOGITIGNORE
bazel-*
*.bazel.lock
**/node_modules
EOGITIGNORE

# Overwrite the .bazelrc to a minimal version not referencing any repos
cat >.bazelrc <<-EOBAZELRC
EOBAZELRC

# Root BUILD
echo "# BLANK" >BUILD.bazel

# Stubs for bazel/ts/defs.bzl
mkdir -p bazel/ts
echo "# BLANK" >bazel/ts/BUILD.bazel
cat >bazel/ts/defs.bzl <<-EOTSDEFS
def ts_config(name, **kwargs):
    pass
def ts_project(name, **kwargs):
    pass
def ts_proto_library(name, **kwargs):
    pass
EOTSDEFS

# Empty BUILD alongside *.MODULE.bazel files
echo "# BLANK" >bazel/include/BUILD.bazel

# MODULE file
cat >MODULE.bazel <<-EOMODULE
module(name = "aspect_orion", version = "0.0.0")

bazel_dep(name = "aspect_bazel_lib", version = "2.10.0")
bazel_dep(name = "aspect_rules_js", version = "2.1.2")

include("//bazel/include:proto.MODULE.bazel")
include("//bazel/include:python.MODULE.bazel")
EOMODULE

# WORKSPACE file
cat >WORKSPACE <<-EOWORKSPACE
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
EOWORKSPACE

# None of the generated files should have changed existing files.
if [[ $(git diff) ]]; then
    echo "$0 should not change any git tracked files"
    git diff
    exit 1
fi
