load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "mock",
    srcs = ["mock.go"],
    importpath = "peridot.resf.org/secparse/db/mock",
    visibility = ["//visibility:public"],
    deps = [
        "//secparse/admin/proto/v1:proto",
        "//secparse/db",
        "//utils",
    ],
)
