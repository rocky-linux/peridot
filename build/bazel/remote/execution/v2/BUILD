load("@io_bazel_rules_go//go:def.bzl", "go_library")
load("@io_bazel_rules_go//proto:def.bzl", "go_proto_library")

package(default_visibility = ["//visibility:public"])

load("@rules_proto//proto:defs.bzl", "proto_library")

licenses(["notice"])

proto_library(
    name = "remoteexecution_proto",
    srcs = ["remote_execution.proto"],
    deps = [
        "//build/bazel/semver:semver_proto",
        "@com_google_protobuf//:duration_proto",
        "@com_google_protobuf//:timestamp_proto",
        "@go_googleapis//google/api:annotations_proto",
        "@go_googleapis//google/longrunning:longrunning_proto",
        "@go_googleapis//google/rpc:status_proto",
    ],
)

# keep
go_proto_library(
    name = "remoteexecution_go_proto",
    compilers = [
        "//:go_apiv2",
        "//:go_grpc",
    ],
    importpath = "bazel.build/remote/execution/v2",
    proto = ":remoteexecution_proto",
    deps = [
        "//build/bazel/semver:semver_go_proto",
        "@go_googleapis//google/api:annotations_go_proto",
        "@go_googleapis//google/longrunning:longrunning_go_proto",
        "@go_googleapis//google/rpc:status_go_proto",
        "@org_golang_google_protobuf//types/known/durationpb",
        "@org_golang_google_protobuf//types/known/timestamppb",
    ],
)

go_library(
    name = "execution",
    embed = [":remoteexecution_go_proto"],
    importpath = "bazel.build/remote/execution/v2",
)
