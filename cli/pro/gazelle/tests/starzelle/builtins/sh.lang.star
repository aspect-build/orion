def prepare(_):
    return aspect.PrepareResult(
        sources = aspect.SourceExtensions(".sh"),
    )

def declare_targets(ctx):
    if len(ctx.sources) == 0:
        ctx.targets.remove("shells")
        ctx.targets.remove("shells-bin")
        return

    ctx.targets.add(
        name = "shells",
        kind = "sh_library",
        attrs = {
            "srcs": [s.path for s in ctx.sources],
        },
    )

    ctx.targets.add(
        name = "shells-bin",
        kind = "sh_binary",
        attrs = {
            "srcs": [s.path for s in ctx.sources],
        },
    )

aspect.register_configure_extension(
    id = "sh-binlib",
    prepare = prepare,
    declare = declare_targets,
)
