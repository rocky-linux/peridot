load("@io_bazel_rules_go//go:def.bzl", "go_library")
load("@io_bazel_rules_go//proto:def.bzl", "go_proto_library")

package(default_visibility = ["//visibility:public"])

load("@rules_proto//proto:defs.bzl", "proto_library")

licenses(["notice"])

proto_library(
    name = "semver_proto",
    srcs = ["semver.proto"],
)

go_proto_library(
    name = "semver_go_proto",
    importpath = "bazel.build/semver",
    proto = ":semver_proto",
)

go_library(
    name = "semver",
    embed = [":semver_go_proto"],
    importpath = "bazel.build/semver",
)
