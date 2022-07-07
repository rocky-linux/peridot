load("@io_bazel_rules_docker//container:container.bzl", "container_pull")
load("//bases:defs.bzl", "ACTIVE_VERSION")

def node_base():
    container_pull(
        name = "node_base_amd64",
        architecture = "amd64",
        registry = "quay.io",
        repository = "peridot/bazel-node",
        tag = ACTIVE_VERSION,
    )
    container_pull(
        name = "node_base_arm64",
        architecture = "arm64",
        registry = "quay.io",
        repository = "peridot/bazel-node",
        tag = ACTIVE_VERSION,
    )
