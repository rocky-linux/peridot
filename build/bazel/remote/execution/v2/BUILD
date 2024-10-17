load("@io_bazel_rules_go//go:def.bzl", "go_library")
load("@io_bazel_rules_go//proto:def.bzl", "go_proto_library")
load("@rules_proto//proto:defs.bzl", "proto_library")

package(default_visibility = ["//visibility:public"])

licenses(["notice"])

proto_library(
    name = "remoteexecution_proto",
    srcs = ["remote_execution.proto"],
    deps = [
        "//build/bazel/semver:semver_proto",
        "@com_google_protobuf//:duration_proto",
        "@com_google_protobuf//:timestamp_proto",
        "@googleapis//google/api:annotations_proto",
        "@googleapis//google/longrunning:operations_proto",
        "@googleapis//google/rpc:status_proto",
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
        "@org_golang_google_genproto//googleapis/longrunning",
        "@org_golang_google_genproto_googleapis_api//annotations",
        "@org_golang_google_genproto_googleapis_rpc//status",
        "@org_golang_google_protobuf//types/known/durationpb",
        "@org_golang_google_protobuf//types/known/timestamppb",
    ],
)

go_library(
    name = "execution",
    embed = [":remoteexecution_go_proto"],
    importpath = "bazel.build/remote/execution/v2",
)
