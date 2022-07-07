load("//rules_byc:defs.bzl", "BYCDEPLOY_OUTS_BASE", "byc_frontend", "container", "peridot_k8s")

package(default_visibility = ["//visibility:public"])

server_deps = [
    "//common/frontend_server",
]

server_entrypoint = "server/index.mjs"

server_srcs = glob(["server/**/*.mjs"])

byc_frontend(
    name = "obsidian",
    srcs = glob([
        "src/**/*.tsx",
        "src/**/*.ts",
    ]),
    entrypoint = "obsidian/ui/src/entrypoint.tsx",
    index_html = "//rules_byc/internal/byc_bundle:index_no_mobile.hbs",
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
        "@npm//@mui/icons-material",
        "@npm//@mui/material",
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
    outs = BYCDEPLOY_OUTS_BASE,
    deps = ["//ci"],
)
