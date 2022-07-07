load("@io_bazel_rules_jsonnet//jsonnet:jsonnet.bzl", "jsonnet_library")

package(default_visibility = ["//visibility:public"])

jsonnet_library(
    name = "ci",
    srcs = glob([
        "**/*.jsonnet",
        "**/*.libsonnet",
    ]),
)
