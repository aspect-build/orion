aspect.register_configure_extension(
    id = "rm-test",
    declare = lambda ctx: ctx.targets.remove("deleteme"),
)
