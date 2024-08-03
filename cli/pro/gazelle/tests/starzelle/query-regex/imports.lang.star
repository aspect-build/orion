aspect.register_rule_kind("x_lib", {
    "From": "@deps-test//my:rules.bzl",
    "MergeableAttrs": ["srcs"],
    "ResolveAttrs": ["deps"],
})

def prepare(_):
    return aspect.PrepareResult(
        # All source files to be processed
        sources = [
            aspect.SourceExtensions(".x"),
        ],
        queries = {
            "imports": aspect.RegexQuery(
                filter = "*.x",
                expression = """import\\s+"(?P<import>[^"]+)\"""",
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
            },
            symbols = [aspect.Symbol(
                id = "/".join([ctx.rel, file.path.removesuffix(".x")]) if ctx.rel else file.path.removesuffix(".x"),
                provider = "x",
            )],
            imports = [
                aspect.Import(
                    id = i.captures["import"],
                    provider = "x",
                    src = file.path,
                )
                for i in file.query_results["imports"]
            ],
        )

aspect.register_configure_extension(
    id = "re-test",
    prepare = prepare,
    declare = declare,
)
