"Maven starlark plugin"

# Directive name and default value from the rules_jvm gazelle plugin
JAVA_MAVEN_INSTALL_FILE = "java_maven_install_file"
DEFAULT_JAVA_MAVEN_INSTALL_FILE = "maven_install.json"

print("hello")

def prepare(_ctx):
    print(_ctx)
    return starzelle.PrepareResult(
        # All source files to be processed
        sources = [
            starzelle.SourceExtensions(DEFAULT_JAVA_MAVEN_INSTALL_FILE),
        ],
        queries = {
            "imports": starzelle.Query(
                grammar = "json",
                filter = DEFAULT_JAVA_MAVEN_INSTALL_FILE,
                query = """
                 (document
                        (object (pair
                            key: (string (string_content) @r1)
                                 (#eq? @r1 "dependency_tree")

                            value: (object (pair
                                key: (string (string_content) @r2)
                                     (#eq? @r2 "dependencies")

                                value: (array
                                    (_) @dep
                                )
                            ))
                        ))
                    )
                """,
            ),
        },
    )

def analyze_source(ctx):
    print(ctx.source)
    print(ctx.source.queries)
    print(ctx.add_symbol)

    for q in ctx.source.queries["imports"]:
        dep = json.decode(q.captures["dep"])
        coord = dep["coord"].split(":", 1)[0]
        if "packages" not in dep:
            continue
        for pkg in dep["packages"]:
            ctx.add_symbol(
                id = pkg,
                provider_type = "java_info",
                label = "@maven//{}".format(coord),
            )

def declare(ctx):
    pass

starzelle.AddLanguagePlugin(
    id = "maven",
    properties = {
        JAVA_MAVEN_INSTALL_FILE: starzelle.Property(
            type = "String",
            default = DEFAULT_JAVA_MAVEN_INSTALL_FILE,
        ),
    },
    prepare = prepare,
    declare = declare,
    analyze = analyze_source,
)
