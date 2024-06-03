# Starzelle BUILD generator

A BUILD generator where plugins implemented in starlakr can be used to generate BUILD files for bazel projects.

## Loading plugins

**WIP**: plugins are loaded via the `${STARZELLE_PLUGINS}/*.lang.star` glob.

If `STARZELLE_PLUGINS` is unspecified `${RUNFILES_DIR}/aspect_silo/cli/pro/gazelle/plugins` is used as a default compatible with gazelle tests within the aspect silo repo.


## TODO:

* change AddKinds API: https://github.com/aspect-build/silo/pull/5625#discussion_r1625063566
