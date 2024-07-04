"""
rules_kotlin support for 'aspect configure'
"""

KT_JVM_LIBRARY = "kt_jvm_library"
KT_JVM_BINARY = "kt_jvm_binary"
RULES_KOTLIN_REPO_NAME = "io_bazel_rules_kotlin"
PROVIDER_NAME = "kt"

LANG_NAME = "kotlin"

starzelle.add_kind(KT_JVM_LIBRARY, {
    "From": "@" + RULES_KOTLIN_REPO_NAME + "//kotlin:jvm.bzl",
    "NonEmptyAttrs": {
        "srcs": True,
    },
    "MergeableAttrs": {
        "srcs": True,
    },
    "ResolveAttrs": {
        "deps": True,
    },
})

starzelle.add_kind(KT_JVM_BINARY, {
    "From": "@" + RULES_KOTLIN_REPO_NAME + "//kotlin:jvm.bzl",
    "NonEmptyAttrs": {
        "srcs": True,
        "main_class": True,
    },
})

def prepare(_):
    return starzelle.PrepareResult(
        # All source files to be processed
        sources = [
            starzelle.SourceExtensions(".kt", ".kts"),
        ],
        queries = {
            "imports": starzelle.Query(
                grammar = "kotlin",
                filter = "*.kt*",
                query = """
                    (source_file
                        (import_list
                            (import_header (identifier) @imp (".*")? @is_star)
                        )
                    )
                """,
            ),
            "package_name": starzelle.Query(
                grammar = "kotlin",
                filter = "*.kt*",
                query = """
                    (source_file
                        (package_header (identifier) @pkg)
                    )
                """,
            ),
            "has_main": starzelle.Query(
                grammar = "kotlin",
                filter = "*.kt*",
                query = """
                    (source_file
                        (function_declaration
                            (simple_identifier) @variable.funcname
                            (#eq? @variable.funcname "main")
                        )
                    )
                """,
            ),
        },
    )

# ctx:
#   rel         string
#   properties  map[string]string
#   sources     []TargetSource
#
# TargetSource:
#   path          string
#   query_results  QueryResults
#
# query_results:
#   [query_key]  bool|string|None
def declare_targets(ctx):
    """
    This function declares targets based on the context.

    Args:
        ctx: The context object.

    Returns:
        a 'DeclareTargetsResult'
    """

    # Every BUILD can have 1 library and multiple binaries
    lib = {
        "srcs": [],
        "packages": [],
        "imports": [],
    }
    bins = []

    for file in ctx.sources:
        pkg = file.query_results["package_name"][0].captures["pkg"] if "package_name" in file.query_results and len(file.query_results["package_name"]) > 0 else None
        if "imports" not in file.query_results:
            print(file.query_results)
        import_paths = [
            i.captures["imp"] if "is_star" in i.captures and i.captures["is_star"] else i.captures["imp"][:i.captures["imp"].rindex(".")]
            for i in file.query_results["imports"]
            if not is_native(i.captures["imp"])
        ]

        # Trim the class name from the import for non-.* imports.
        # Convert to TargetImport, exclude native imports
        imports = [
            starzelle.Import(
                id = i,
                provider = PROVIDER_NAME,
                src = file.path,
            )
            for i in import_paths
        ]

        if len(file.query_results["has_main"]) > 0:
            bins.append({
                "src": file.path,
                "imports": imports,
                "package": starzelle.Symbol(
                    id = pkg,
                    provider = PROVIDER_NAME,
                ) if pkg else None,
            })
        else:
            lib["srcs"].append(file.path)
            lib["imports"].extend(imports)
            if pkg:
                lib["packages"].append(starzelle.Symbol(
                    id = pkg,
                    provider = PROVIDER_NAME,
                ))

    lib_name = path.base(ctx.rel) if ctx.rel else ctx.repo_name
    if len(lib["srcs"]) > 0:
        ctx.targets.add(
            name = lib_name,
            kind = KT_JVM_LIBRARY,
            attrs = {
                "srcs": lib["srcs"],
            },
            symbols = lib["packages"],
            imports = lib["imports"],
        )
    else:
        ctx.targets.remove(lib_name)

    for bin in bins:
        no_ext = bin["src"].removesuffix(path.ext(bin["src"]))

        ctx.targets.add(
            name = no_ext.lower() + "_bin",
            kind = KT_JVM_BINARY,
            attrs = {
                "srcs": [bin["src"]],
                "main_class": (bin["package"].id + "." + no_ext) if bin["package"] else no_ext,
            },
            symbols = [bin["package"]] if bin["package"] else [],
            imports = bin["imports"],
        )

NATIVE_LIBS = [
    "kotlin",
    "kotlinx",

    # Java, see rules_jvm gazelle plugin:
    # https://github.com/bazel-contrib/rules_jvm/blob/v0.24.0/java/gazelle/private/java/java.go#L28-L147
    "java",
    "javax",
    "com.sun",
    "jdk",
    "netscape.javascript",
    "org.ietf.jgss",
    "org.jcp.xml.dsig.internal",
    "org.w3c.dom",
    "org.xml.sax",
    "sun",
]

def is_native(imp):
    for lib in NATIVE_LIBS:
        if imp == lib or imp.startswith(lib + "."):
            return True

    return False

starzelle.add_plugin(
    id = LANG_NAME,
    properties = {},
    prepare = prepare,
    declare = declare_targets,
)
