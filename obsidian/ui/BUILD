load("//rules_resf:defs.bzl", "RESFDEPLOY_OUTS_BASE", "container", "peridot_k8s", "resf_frontend")

package(default_visibility = ["//visibility:public"])

server_deps = [
    "//common/frontend_server",
]

server_entrypoint = "server/index.mjs"

server_srcs = glob(["server/**/*.mjs"])

resf_frontend(
    name = "obsidian",
    srcs = glob([
        "src/**/*.tsx",
        "src/**/*.ts",
    ]),
    entrypoint = "obsidian/ui/src/entrypoint.tsx",
    index_html = "//rules_resf/internal/resf_bundle:index_no_mobile.hbs",
    server_deps = server_deps,
    server_entrypoint = server_entrypoint,
    server_srcs = server_srcs,
    title = "Obsidian ID Service",
    deps = [
        "//common/mui",
        "//common/ui",
        "//dotui",
        "//obsidian/proto/v1:client_typescript",
        "//tailwind:css",
        "@npm//@emotion/unitless",
        "@npm//@mui/icons-material",
        "@npm//@mui/material",
        "@npm//@mui/styles",
        "@npm//@mui/system",
        "@npm//await-to-js",
        "@npm//react-router",
        "@npm//react-router-dom",
        "@npm//react-social-login-buttons",
    ],
)

container(
    base = "//bases/bazel/node",
    files = [
        ":obsidian.bundle",
    ],
    frontend = True,
    image_name = "obsidian-frontend",
    server_entrypoint = server_entrypoint,
    server_files = server_srcs + server_deps,
)

peridot_k8s(
    name = "obsidian-frontend",
    src = "deploy.jsonnet",
    outs = RESFDEPLOY_OUTS_BASE,
    chart_yaml = "Chart.yaml",
    values_yaml = "values.yaml",
    deps = ["//ci"],
)
