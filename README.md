# Orion

The Aspect CLI Orion feature in a standalone repository.

## Syncing

This repository is synced from the internal Aspect monorepo.

A subset of files which can not be synced are manually maintained, see the
[manual-files branch](https://github.com/aspect-build/orion/tree/manual-files)
which contains commit(s) required *BEFORE* applying the sync from the Aspect monorepo.

Steps:
1. start with the [manual-files branch](https://github.com/aspect-build/orion/tree/manual-files)
2. [git-filter-repo](https://github.com/newren/git-filter-repo) from the Aspect monorepo (see monorepo for syncing setup) to populate this repository.
3. Ensure `bazel test //...` passes. If tests do not pass:
  * ensure all required files are included in the `git-filter-repo` setup
  * ensure manually maintained files are up to date such as WORKSPACE

## Targets

* `//cli/pro/gazelle/host` - the [gazelle language](https://pkg.go.dev/github.com/bazelbuild/bazel-gazelle/language) implementation of Orion.
* `//cli/pro/gazelle/host:gazelle_host_binary` a [gazelle binary](https://github.com/bazel-contrib/bazel-gazelle/blob/master/extend.md#gazelle_binary) with Orion included.
* See `bazel test //...` for test targets
