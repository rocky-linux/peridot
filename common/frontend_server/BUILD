load("@raw_ts_library//:index.bzl", "raw_ts_library")

raw_ts_library(
    name = "frontend_server",
    srcs = [
        "index.mjs",
        "upstream.mjs",
    ],
    module_name = "common/frontend_server",
    visibility = ["//visibility:public"],
    deps = [
        "@npm//body-parser",
        "@npm//connect-history-api-fallback",
        "@npm//cookie-parser",
        "@npm//evil-dns",
        "@npm//express",
        "@npm//express-openid-connect",
        "@npm//hbs",
        "@npm//helmet",
        "@npm//http-proxy-middleware",
        "@npm//webpack",
        "@npm//webpack-dev-middleware",
        "@npm//webpack-hot-middleware",
        "@npm//webpack-mild-compile",
    ],
)
