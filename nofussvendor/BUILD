load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "nofussvendor_lib",
    srcs = ["main.go"],
    importpath = "peridot.resf.org/nofussvendor",
    visibility = ["//visibility:private"],
    deps = [
        "//vendor/bazel.build/protobuf",
        "//vendor/github.com/sirupsen/logrus",
        "@org_golang_google_protobuf//proto:go_default_library",
    ],
)

go_binary(
    name = "nofussvendor",
    embed = [":nofussvendor_lib"],
    visibility = ["//visibility:public"],
)
