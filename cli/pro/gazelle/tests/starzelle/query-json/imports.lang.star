aspect.register_rule_kind("x_lib", {
    "From": "@deps-test//my:rules.bzl",
    "MergeableAttrs": ["srcs"],
    "ResolveAttrs": ["deps"],
})

def prepare(_):
    return aspect.PrepareResult(
        # All source files to be processed
        sources = [
            aspect.SourceExtensions(".json"),
        ],
        queries = {
            "imports": aspect.JsonQuery(
                filter = "*.json",
                query = ".imports[]?",
            ),
        },
    )

def declare(ctx):
    for file in ctx.sources:
        ctx.targets.add(
            name = file.path[:file.path.rindex(".")] + "_lib",
            kind = "x_lib",
            attrs = {
                "srcs": [file.path],
                "deps": [
                    aspect.Import(
                        id = i,
                        provider = "x",
                        src = file.path,
                    )
                    for i in file.query_results["imports"]
                    if i
                ],
            },
            symbols = [aspect.Symbol(
                id = "/".join([ctx.rel, file.path.removesuffix(".json")]) if ctx.rel else file.path.removesuffix(".json"),
                provider = "x",
            )],
        )

aspect.register_configure_extension(
    id = "jsonq-test",
    prepare = prepare,
    declare = declare,
)
