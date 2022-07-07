ModuleNameRoot = provider(
    doc = "provides module_name and module_root to other rules",
    fields = {
        "module_name": "",
        "module_root": "",
    },
)

# From bazelbuild/rules_typescript
# https://github.com/bazelbuild/rules_typescript/blob/2312a8507090182d6565a4b072eb7893b20bce0b/internal/common/compilation.bzl
def ts_providers_dict_to_struct(d):
    for key, value in d.items():
        if key != "output_groups" and type(value) == type({}):
            d[key] = struct(**value)
    return struct(**d)

def _raw_ts_library(ctx):
    return ts_providers_dict_to_struct({
        "files": depset(ctx.files.srcs + ctx.files.deps),
        "module_name": ctx.attr.module_name,
        "module_root": ctx.attr.module_root,
        "raw_ts": True,
    })

raw_ts_library = rule(
    attrs = {
        "srcs": attr.label_list(
            allow_files = [".js", ".jsx", ".mjs", ".ts", ".tsx", ".less", ".scss", ".css"],
        ),
        "deps": attr.label_list(
            allow_files = True,
        ),
        "module_name": attr.string(),
        "module_root": attr.string(),
    },
    implementation = _raw_ts_library,
)
