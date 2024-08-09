aspect.register_rule_kind("x_lib", {
    "From": "@deps-test//my:rules.bzl",
})

def prepare(_):
    return aspect.PrepareResult(
        # TODO: need source otherwise `declare` is never invoked
        sources = [
            aspect.SourceExtensions(".conf"),
        ],
    )

def declare(ctx):
    ctx.targets.add(
        name = "a",
        kind = "x_lib",
        symbols = [
            aspect.Symbol(
                id = path.join(ctx.rel, "generated"),
                provider = "js",
            ),
        ],
    )

aspect.register_configure_extension(
    id = "js-imports-test",
    prepare = prepare,
    declare = declare,
)
