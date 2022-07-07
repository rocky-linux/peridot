load("@io_bazel_rules_docker//container:container.bzl", "container_pull")
load("//bases:defs.bzl", "ACTIVE_VERSION")

def go_base():
    container_pull(
        name = "go_base_amd64",
        architecture = "amd64",
        registry = "quay.io",
        repository = "peridot/bazel-go",
        tag = ACTIVE_VERSION,
    )
    container_pull(
        name = "go_base_arm64",
        architecture = "arm64",
        registry = "quay.io",
        repository = "peridot/bazel-go",
        tag = ACTIVE_VERSION,
    )
    container_pull(
        name = "go_base_s390x",
        architecture = "s390x",
        registry = "quay.io",
        repository = "peridot/bazel-go",
        tag = ACTIVE_VERSION,
    )
    container_pull(
        name = "go_base_ppc64le",
        architecture = "ppc64le",
        registry = "quay.io",
        repository = "peridot/bazel-go",
        tag = ACTIVE_VERSION,
    )
