load("@raw_ts_library//:index.bzl", "raw_ts_library")

package(default_visibility = ["//visibility:public"])

raw_ts_library(
    name = "dotui",
    srcs = glob([
        "*.ts",
        "*.tsx",
    ]),
    module_name = "dotui",
    deps = [
        "@npm//react",
        "@npm//react-dom",
    ],
)
