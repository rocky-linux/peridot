load("@raw_ts_library//:index.bzl", "raw_ts_library")

package(default_visibility = ["//visibility:public"])

raw_ts_library(
    name = "ui",
    srcs = glob([
        "*.ts",
        "*.tsx",
        "*.less",
    ]),
    module_name = "common/ui",
    deps = [
        "@npm//ansi-regex",
        "@npm//classnames",
        "@npm//path-to-regexp",
        "@npm//popper.js",
        "@npm//strip-ansi",
    ],
)
