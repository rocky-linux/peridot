load("@io_bazel_rules_go//go:def.bzl", "go_library")
load("@com_google_protobuf//:protobuf.bzl", "py_proto_library")
load("@io_bazel_rules_go//proto:def.bzl", "go_proto_library")
load("@rules_cc//cc:defs.bzl", "cc_library", "cc_proto_library")
load("@rules_java//java:defs.bzl", "java_proto_library")
load("@rules_proto//proto:defs.bzl", "proto_library")

package(
    default_visibility =
        ["//visibility:public"],
)

proto_library(
    name = "validate_proto",
    srcs = ["validate.proto"],
    deps = [
        "@com_google_protobuf//:descriptor_proto",
        "@com_google_protobuf//:duration_proto",
        "@com_google_protobuf//:timestamp_proto",
    ],
)

cc_proto_library(
    name = "validate_cc",
    deps = [":validate_proto"],
)

py_proto_library(
    name = "validate_py",
    srcs = ["validate.proto"],
    deps = ["@com_google_protobuf//:protobuf_python"],
)

go_proto_library(
    name = "go_default_library",
    importpath = "github.com/envoyproxy/protoc-gen-validate/validate",
    proto = ":validate_proto",
)

cc_library(
    name = "cc_validate",
    hdrs = ["validate.h"],
)

java_proto_library(
    name = "validate_java",
    deps = [":validate_proto"],
)

go_library(
    name = "validate",
    srcs = [
        "validate.h",
        "validate.pb.go",
    ],
    importmap = "peridot.resf.org/vendor/github.com/envoyproxy/protoc-gen-validate/validate",
    importpath = "github.com/envoyproxy/protoc-gen-validate/validate",
    deps = [
        "@org_golang_google_protobuf//reflect/protoreflect:go_default_library",
        "@org_golang_google_protobuf//runtime/protoimpl:go_default_library",
        "@org_golang_google_protobuf//types/descriptorpb:go_default_library",
        "@org_golang_google_protobuf//types/known/durationpb:go_default_library",
        "@org_golang_google_protobuf//types/known/timestamppb:go_default_library",
    ],
)
