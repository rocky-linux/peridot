load("@build_bazel_rules_nodejs//:index.bzl", "nodejs_binary")
load(":resf_bundle.bzl", "WEBPACK_DATA")

package(default_visibility = ["//rules_resf/internal:__subpackages__"])

exports_files([
    "assembler.js",
    "webpack.config.js",
    "tsconfig.json",
    "babel.config.js",
    "tailwind.config.js",
    "tailwind.config.nopreflight.js",
    "package.json",
    "yarn.lock",
    "index.hbs",
    "index_no_mobile.hbs",
    "index_internal.hbs",
    "resf_bundle.bzl",
    "run_child.mjs",
    "packserver.bash",
])

nodejs_binary(
    name = "run_child",
    data = [":run_child.mjs"],
    entry_point = ":run_child.mjs",
    link_workspace_root = True,
    templated_args = ["--bazel_patch_module_resolver"],
    visibility = ["//visibility:public"],
)

nodejs_binary(
    name = "webpack",
    # Allow --define=ROLLUP_BUNDLE_FIXED_CHUNK_NAMES=1 to be passed as
    # process.env so we can give predictable chunk names in CI
    configuration_env_vars = [""],
    data = WEBPACK_DATA,
    entry_point = "@npm//:node_modules/webpack-cli/bin/cli.js",
    link_workspace_root = True,
    templated_args = [
        "--bazel_patch_module_resolver",
        "--node_options=--max_old_space_size=1200000000",
        "--node_options=--stack-size=650000",
    ],
    visibility = ["//visibility:public"],
)
