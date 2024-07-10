"Maven starlark plugin"

# Directive name and default value from the rules_jvm gazelle plugin
JAVA_MAVEN_INSTALL_FILE = "java_maven_install_file"
DEFAULT_JAVA_MAVEN_INSTALL_FILE = "maven_install.json"

def prepare(ctx):
    return aspect.PrepareResult(
        # All source files to be processed
        sources = [
            aspect.SourceExtensions(ctx.properties[JAVA_MAVEN_INSTALL_FILE]),
        ],
        queries = {
            "imports": aspect.Query(
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
    for q in ctx.source.query_results["imports"]:
        dep = json.decode(q.captures["dep"])

        if "packages" not in dep:
            continue

        coord = dep["coord"].rsplit(":", 1)[0].replace(".", "_").replace(":", "_")

        for pkg in dep["packages"]:
            ctx.add_symbol(
                id = pkg,
                provider_type = "java_info",
                label = aspect.Label(
                    repo = "maven",
                    name = coord,
                ),
            )

aspect.register_configure_extension(
    id = "maven",
    properties = {
        JAVA_MAVEN_INSTALL_FILE: aspect.Property(
            type = "String",
            default = DEFAULT_JAVA_MAVEN_INSTALL_FILE,
        ),
    },
    prepare = prepare,
    analyze = analyze_source,
)
