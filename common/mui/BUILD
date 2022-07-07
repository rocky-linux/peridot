load("@raw_ts_library//:index.bzl", "raw_ts_library")

package(default_visibility = ["//visibility:public"])

raw_ts_library(
    name = "mui",
    srcs = glob([
        "*.ts",
        "*.tsx",
        "*.less",
    ]),
    module_name = "common/mui",
    deps = [
        "@npm//@emotion/react",
        "@npm//@emotion/styled",
        "@npm//@mui/material",
        "@npm//@mui/styles",
    ],
)
