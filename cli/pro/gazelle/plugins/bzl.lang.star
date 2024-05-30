BZL_LIBRARY = "bzl_library"

LANG_NAME = "starlark"
BZL_EXT = ".bzl"

LANG_RULES = {
    BZL_LIBRARY: {
        "From": "@bazel_skylib//:bzl_library.bzl",
    },
}

def Prepare(_):
    return starzelle.PrepareResult(
        sources = [
            starzelle.SourceExtensions(".bzl"),
        ],
        queries = {
            "loads": starzelle.Query(
                query = """(module
                    (expression_statement
                        (call 
                            function: (identifier) @id
                            arguments: (argument_list
                                (string) @path
                                (string)
                            )
                        )
                        (#eq? @id "load")
                    )
                )""",
            ),
        },
    )

def DeclareTargets(ctx):
    # TODO
    # Loop through the existing bzl_library targets in this package and
    # delete any that are no longer needed.
    for file in ctx.sources:
        label = file.path.removesuffix(".bzl").replace("/", "_")
        file_pkg = path.dirname(file.path)

        loads = [ld.captures["path"].strip("\"") for ld in file.query_results["loads"]]
        loads = [ld.removeprefix("//").replace(":", "/") if ld.startswith("//") else path.join(file_pkg, ld.replaceprefix(":")) for ld in loads]
        loads = [ld.strip("/") for ld in loads]

        ctx.targets.add(
            kind = BZL_LIBRARY,
            name = label,
            attrs = {
                "srcs": [file.path],
                "visibility": ["//visibility:public"],
            },
            # TODO
            # load("@bazel_tools//tools/build_defs/repo:http.bzl")
            # Note that the Go extension has a special case for it:
            # if impLabel.Repo == "bazel_tools" {
            # // The @bazel_tools repo is tricky because it is a part of the "shipped
            # // with bazel" core library for interacting with the outside world.
            imports = [
                starzelle.Import(
                    id = ld,
                    src = file.path,
                )
                for ld in loads
            ],
            symbols = [
                starzelle.Symbol(
                    id = file.path,
                    label = label,
                ),
            ],
        )
    return {}

starzelle.AddLanguagePlugin(
    id = LANG_NAME,
    properties = {},
    rules = LANG_RULES,
    prepare = Prepare,
    declare = DeclareTargets,
)
