aspect.register_rule_kind("x_lib", {
    "From": "@deps-test//my:rules.bzl",
})

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
    declare = declare,
)
