load("@io_bazel_rules_docker//container:container.bzl", "container_image")

container_image(
    name = "node",
    base = select({
        "//platforms:arm64": "@node_base_arm64//image",
        "//platforms:x86_64": "@node_base_amd64//image",
    }),
    visibility = ["//visibility:public"],
)
