def prepare(_):
    return aspect.PrepareResult(
        sources = [
            aspect.SourceGlobs("**/*.txt"),
        ],
    )

def declare_targets(ctx):
    ctx.targets.add(
        name = ctx.rel,
        kind = "copy_to_bin",
        attrs = {
            "srcs": [s.path for s in ctx.sources],
        },
    )

aspect.register_configure_extension(
    id = "copy-txt",
    prepare = prepare,
    declare = declare_targets,
)
