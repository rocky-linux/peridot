load("@build_bazel_rules_nodejs//:providers.bzl", "JSEcmaScriptModuleInfo", "JSModuleInfo", "JSNamedModuleInfo", "NpmPackageInfo", "node_modules_aspect", "run_node")
load("@build_bazel_rules_nodejs//internal/linker:link_node_modules.bzl", "module_mappings_aspect")
load("@build_bazel_rules_nodejs//:index.bzl", "nodejs_binary")

def _trim_package_node_modules(package_name):
    # trim a package name down to its path prior to a node_modules
    # segment. 'foo/node_modules/bar' would become 'foo' and
    # 'node_modules/bar' would become ''
    segments = []
    for n in package_name.split("/"):
        if n == "node_modules":
            break
        segments += [n]
    return "/".join(segments)

# This function is similar but slightly different than _compute_node_modules_root
# in /internal/node/node.bzl. TODO(gregmagolan): consolidate these functions
def _compute_node_modules_root(ctx):
    """Computes the node_modules root from the node_modules and deps attributes.
    Args:
      ctx: the skylark execution context
    Returns:
      The node_modules root as a string
    """
    node_modules_root = None
    if ctx.files.node_modules:
        # ctx.files.node_modules is not an empty list
        node_modules_root = "/".join([f for f in [
            ctx.attr.node_modules.label.workspace_root,
            _trim_package_node_modules(ctx.attr.node_modules.label.package),
            "node_modules",
        ] if f])
    for d in ctx.attr.deps:
        if NpmPackageInfo in d:
            possible_root = "/".join(["external", d[NpmPackageInfo].workspace, "node_modules"])
            if not node_modules_root:
                node_modules_root = possible_root
            elif node_modules_root != possible_root:
                fail("All npm dependencies need to come from a single workspace. Found '%s' and '%s'." % (node_modules_root, possible_root))
    if not node_modules_root:
        # there are no fine grained deps and the node_modules attribute is an empty filegroup
        # but we still need a node_modules_root even if its empty
        node_modules_root = "@npm//:node_modules"
    return node_modules_root

def collect_ts_sources(ctx):
    non_rerooted_files = [d for d in ctx.files.deps]
    if hasattr(ctx.attr, "srcs"):
        non_rerooted_files += ctx.files.srcs

    rerooted_files = []
    for file in non_rerooted_files:
        if file.is_directory:
            rerooted_files += [file]
            continue

        path = file.short_path
        if (path.startswith("../")):
            path = "external/" + path[3:]

        rerooted_file = ctx.actions.declare_file(
            "%s" % (
                path.replace(".closure.js", ".ts").replace(ctx.label.package + "/", ""),
            ),
        )

        # Cheap way to create an action that copies a file
        # TODO(alexeagle): discuss with Bazel team how we can do something like
        # runfiles to create a re-rooted tree. This has performance implications.
        ctx.actions.expand_template(
            output = rerooted_file,
            template = file,
            substitutions = {},
        )
        rerooted_files += [rerooted_file]

    #TODO(mrmeku): we should include the files and closure_js_library contents too
    return depset(direct = rerooted_files)

def _filter_js_inputs(all_inputs):
    # Note: make sure that "all_inputs" is not a depset.
    # Iterating over a depset is deprecated!
    return [
        f
        for f in all_inputs
        # We also need to include ".map" files as these can be read by
        # the "rollup-plugin-sourcemaps" plugin.
        if f.path.endswith(".js") or f.path.endswith(".jsx") or f.path.endswith(".json") or f.path.endswith(".map") or f.path.endswith(".ts") or f.path.endswith(".tsx") or f.path.endswith(".css")
    ]

def get_inputs(ctx, config, dep_files_attr, dep_attr, index_html = None, extra = []):
    direct_inputs = [] + extra + dep_files_attr
    if config:
        direct_inputs.append(config)
    if index_html:
        direct_inputs.append(index_html)

    if ctx.files.node_modules:
        direct_inputs += _filter_js_inputs(ctx.files.node_modules)

    # Also include files from npm fine grained deps as inputs.
    # These deps are identified by the NpmPackageInfo provider.
    for d in dep_attr:
        if NpmPackageInfo in d:
            # Note: we can't avoid calling .to_list() on sources
            direct_inputs += _filter_js_inputs(d[NpmPackageInfo].sources.to_list())
        else:
            direct_inputs += _filter_js_inputs(d.files.to_list())

    if ctx.file.license_banner:
        direct_inputs += [ctx.file.license_banner]

    return direct_inputs

def run_webpack(ctx, sources, config, output, map_output = None, direct_inputs = None, index_html = None):
    args = ctx.actions.args()
    args.add_all(["--config", config.path])
    args.add_all(["--output-path", output.path])
    # args.add_all(["--silent"])

    outputs = [output]

    ctx.actions.run(
        progress_message = "Bundling TypeScript %s [webpack]" % ctx.attr.name,
        executable = ctx.executable._webpack,
        inputs = depset(direct_inputs, transitive = [sources, depset(ctx.files.srcs)]),
        outputs = outputs,
        arguments = [args],
        env = {
            "NODE_ENV": "production",
            "BABEL_ENV": "production",
            "TAILWIND_DISABLE_TOUCH": "true",
        },
    )

def write_index_html(ctx, filename = "index.hbs", output = None, index_html = None):
    html = ctx.actions.declare_file(filename) if not output else output

    ctx.actions.expand_template(
        output = html,
        template = index_html,
        substitutions = {
            "TMPL_name": ctx.attr.title if ctx.attr.title else "Peridot",
            "TMPL_bundle": ctx.label.name,
            "TMPL_prefix": ctx.attr.prefix,
        },
    )

    return html

def write_webpack_config(ctx, plugins = [], root_dir = None, filename = "_%s.webpack.config.js", output_format = "iife", additional_entrypoints = [], index_html = None):
    """Generate a rollup config file.
    This is also used by the ng_rollup_bundle and ng_package rules in @angular/bazel.
    Args:
      ctx: Bazel rule execution context
      plugins: extra plugins (defaults to [])
               See the ng_rollup_bundle in @angular/bazel for example of usage.
      root_dir: root directory for module resolution (defaults to None)
      filename: output filename pattern (defaults to `_%s.rollup.conf.js`)
      output_format: passed to rollup output.format option, e.g. "umd"
      additional_entrypoints: additional entry points for code splitting
    Returns:
      The rollup config file. See https://rollupjs.org/guide/en#configuration-files
    """
    config = ctx.actions.declare_file(filename % ctx.label.name)

    # build_file_path includes the BUILD.bazel file, transform here to only include the dirname
    build_file_dirname = "/".join(ctx.build_file_path.split("/")[:-1])

    entrypoints = [ctx.attr.entrypoint] + additional_entrypoints

    mappings = dict()
    all_deps = ctx.attr.deps
    for dep in all_deps:
        if hasattr(dep, "module_name"):
            mappings[dep.module_name] = dep.label.package

    if not root_dir:
        # This must be .es6 to match collect_es6_sources.bzl
        root_dir = "/".join([ctx.bin_dir.path, build_file_dirname, ctx.label.name + ".es6"])

    node_modules_root = _compute_node_modules_root(ctx)
    is_default_node_modules = False
    if node_modules_root == "node_modules" and ctx.attr.node_modules.label.package == "" and ctx.attr.node_modules.label.name == "node_modules_none":
        is_default_node_modules = True

    direct_inputs = get_inputs(ctx, config, ctx.files.deps, ctx.attr.deps, index_html = index_html, extra = [ctx.file._tsconfig, ctx.file._babel_config, ctx.file.tailwind_config, ctx.file._base_tailwind_config])

    input_paths = []
    for input in direct_inputs:
        path = input.path
        if not "node_modules" in path:
            input_paths.append(path)

    ctx.actions.expand_template(
        output = config,
        template = ctx.file._webpack_config_tmpl,
        substitutions = {
            "TMPL_additional_plugins": ",\n".join(plugins),
            "TMPL_banner_file": "\"%s\"" % ctx.file.license_banner.path if ctx.file.license_banner else "undefined",
            "TMPL_global_name": ctx.attr.global_name if ctx.attr.global_name else ctx.label.name,
            "TMPL_no_suffix_frontend": "true" if ctx.attr.no_suffix_frontend else "false",
            "TMPL_inputs": ",".join(["\"%s\"" % e for e in entrypoints]),
            "TMPL_module_mappings": str(mappings),
            "TMPL_output_format": output_format,
            "TMPL_indexHtml": index_html.short_path if index_html != None else "null",
            "TMPL_target": str(ctx.label),
            "TMPL_title": ctx.attr.title if ctx.attr.title else "Peridot",
            "TMPL_body_script": ctx.attr.script,
            "TMPL_head_style": ctx.attr.style,
            "TMPL_typekit": ctx.attr.typekit,
            "TMPL_api_url": ctx.var.API_URL if "API_URL" in ctx.var else "",
            "TMPL_api_key": ctx.var.API_KEY if "API_KEY" in ctx.var else "",
            "TMPL_tailwind_config": ctx.file.tailwind_config.path,
        },
    )

    return dict({
        "webpack": config,
        "inputs": direct_inputs,
    })

def packserver(ctx, webpack_config, webpack_inputs):
    name = "{}.server".format(ctx.attr.name)
    direct_inputs = get_inputs(ctx, None, ctx.files.server_deps, ctx.attr.server_deps)
    direct_inputs_srcs = get_inputs(ctx, webpack_config, ctx.files.deps, ctx.attr.deps)

    """args = ctx.actions.args()
    args.add(ctx.file.server_entrypoint.short_path)
    args.add(webpack_config.path)

    run_node(
        ctx,
        arguments = [args],
        progress_message = "Packing frontend server %s" % ctx.attr.name,
        executable = "_run_child",
        inputs = direct_inputs + webpack_inputs + ctx.files.server_srcs + [webpack_config],
        outputs = [server],
    )"""
    node_modules_root = _compute_node_modules_root(ctx)

    out_file = ctx.actions.declare_file(name + ".bash")
    ctx.actions.expand_template(
        template = ctx.file._packserver,
        output = out_file,
        substitutions = {
            "TMPL_run_child": ctx.file._run_child_script.short_path,
            "TMPL_node": ctx.executable._node.short_path,
            "TMPL_entrypoint": ctx.file.server_entrypoint.short_path,
            "TMPL_webpack_config_path": webpack_config.short_path,
        },
        is_executable = True,
    )

    runfiles = ctx.runfiles().merge(ctx.attr._node_bash_runfiles[DefaultInfo].default_runfiles)
    runfiles = runfiles.merge(ctx.attr._node[DefaultInfo].default_runfiles)
    runfiles = runfiles.merge(ctx.runfiles(
        transitive_files = ctx.attr._node.files,
        files = direct_inputs + direct_inputs_srcs + ctx.files.srcs + webpack_inputs + ctx.files._webpack_data + ctx.files.server_srcs + collect_ts_sources(ctx).to_list() + [webpack_config, ctx.file._run_child_script] + ctx.files._node + ctx.files._node_files,
    ))

    return [DefaultInfo(
        files = depset([out_file]),
        runfiles = runfiles,
        executable = out_file,
    )]

def _resf_bundle(ctx):
    index_html = ctx.file.index_html
    config = write_webpack_config(ctx, index_html = index_html)
    webpack_config = config["webpack"]
    direct_inputs = config["inputs"]

    # Generate the bundles
    if ctx.attr.build:
        ui = ctx.actions.declare_directory("{}.ui".format(ctx.attr.name))
        run_webpack(ctx, collect_ts_sources(ctx), webpack_config, ui, direct_inputs = direct_inputs, index_html = index_html)

        files = [ui]
        output_group = OutputGroupInfo(
            ui = depset(files),
        )

        runfiles = ctx.runfiles(files)

        return [
            DefaultInfo(files = depset(files), runfiles = runfiles),
            output_group,
        ]
    else:
        return packserver(ctx, webpack_config, direct_inputs)

WEBPACK_DATA = [
    "//rules_resf/internal/resf_bundle:babel.config.js",
    "//rules_resf/internal/resf_bundle:tailwind.config.js",
    "//rules_resf/internal/resf_bundle:index.hbs",
    "//rules_resf/internal/resf_bundle:tsconfig.json",
    "@npm//@babel/plugin-transform-modules-commonjs",
    "@npm//@babel/preset-env",
    "@npm//@babel/preset-react",
    "@npm//@babel/preset-typescript",
    "@npm//@tailwindcss/forms",
    "@npm//autoprefixer",
    "@npm//babel-loader",
    "@npm//babel-plugin-import",
    "@npm//compression-webpack-plugin",
    "@npm//css-loader",
    "@npm//error-stack-parser",
    "@npm//file-loader",
    "@npm//glob",
    "@npm//html-webpack-plugin",
    "@npm//fs-extra",
    "@npm//mini-css-extract-plugin",
    "@npm//native-url",
    "@npm//optimize-css-assets-webpack-plugin",
    "@npm//postcss",
    "@npm//postcss-loader",
    "@npm//purgecss-webpack-plugin",
    "@npm//stackframe",
    "@npm//strip-ansi",
    "@npm//style-loader",
    "@npm//tailwindcss",
    "@npm//terser-webpack-plugin",
    "@npm//@pmmmwh/react-refresh-webpack-plugin",
    "@npm//react-refresh",
    "@npm//type-fest",
    "@npm//webpack",
    "@npm//webpack-cli",
    "@npm//webpack-mild-compile",
    "@npm//ansi-html-community",
    "@npm//core-js-pure",
]

resf_bundle_ATTRS = {
    "title": attr.string(),
    "script": attr.string(
        default = "",
    ),
    "style": attr.string(
        default = "",
    ),
    "typekit": attr.string(
        #default = "https://use.typekit.net/fjm0njo.css",
        default = "https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&display=swap",
    ),
    "index_html": attr.label(
        default = Label("//rules_resf/internal/resf_bundle:index.hbs"),
        allow_single_file = True,
    ),
    "prefix": attr.string(
        default = "/",
    ),
    "srcs": attr.label_list(
        allow_files = True,
        default = [],
    ),
    "deps": attr.label_list(
        aspects = [module_mappings_aspect, node_modules_aspect],
        allow_files = True,
    ),
    "node_modules": attr.label_list(
        allow_files = True,
    ),
    "license_banner": attr.label(
        allow_single_file = True,
    ),
    "global_name": attr.string(),
    "no_suffix_frontend": attr.bool(
        default = False,
    ),
    "build": attr.bool(
        default = True,
    ),
    "server_entrypoint": attr.label(
        allow_single_file = True,
        mandatory = True,
    ),
    "server_srcs": attr.label_list(
        allow_files = True,
        mandatory = True,
    ),
    "server_deps": attr.label_list(
        aspects = [module_mappings_aspect, node_modules_aspect],
    ),
    "_webpack_config_tmpl": attr.label(
        default = Label("//rules_resf/internal/resf_bundle:webpack.config.js"),
        allow_single_file = True,
    ),
    "_webpack": attr.label(
        default = Label("//rules_resf/internal/resf_bundle:webpack"),
        executable = True,
        cfg = "host",
        allow_files = True,
    ),
    "_webpack_data": attr.label_list(
        default = WEBPACK_DATA,
        allow_files = True,
    ),
    "_node": attr.label(
        default = Label("@nodejs//:node_bin"),
        allow_single_file = True,
        executable = True,
        cfg = "host",
    ),
    "_node_files": attr.label_list(
        default = [Label("@nodejs//:node_files")],
        allow_files = True,
    ),
    "_run_child": attr.label(
        default = Label("//rules_resf/internal/resf_bundle:run_child"),
        executable = True,
        cfg = "host",
        allow_files = True,
    ),
    "_run_child_script": attr.label(
        default = Label("//rules_resf/internal/resf_bundle:run_child.mjs"),
        allow_single_file = True,
    ),
    "_packserver": attr.label(
        default = Label("//rules_resf/internal/resf_bundle:packserver.bash"),
        allow_single_file = True,
    ),
    "_tsconfig": attr.label(
        default = Label("//rules_resf/internal/resf_bundle:tsconfig.json"),
        allow_single_file = True,
    ),
    "entrypoint": attr.string(
        default = "src/entrypoint.tsx",
    ),
    "entry_point": attr.string(
        mandatory = False,
    ),
    "_babel_config": attr.label(
        default = Label("//rules_resf/internal/resf_bundle:babel.config.js"),
        allow_single_file = True,
    ),
    "tailwind_config": attr.label(
        default = Label("//rules_resf/internal/resf_bundle:tailwind.config.js"),
        allow_single_file = True,
    ),
    "_base_tailwind_config": attr.label(
        default = Label("//rules_resf/internal/resf_bundle:tailwind.config.js"),
        allow_single_file = True,
    ),
    "_bash_runfiles": attr.label(
        default = Label("@bazel_tools//tools/bash/runfiles"),
    ),
    "_node_bash_runfiles": attr.label(
        default = Label("@build_bazel_rules_nodejs//third_party/github.com/bazelbuild/bazel/tools/bash/runfiles"),
    ),
}

resf_bundle = rule(
    implementation = _resf_bundle,
    attrs = resf_bundle_ATTRS,
)

resf_bundle_run = rule(
    implementation = _resf_bundle,
    attrs = resf_bundle_ATTRS,
    executable = True,
)
