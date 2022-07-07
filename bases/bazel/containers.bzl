load("@io_bazel_rules_docker//container:container.bzl", "container_pull")
load("//bases/bazel/node:defs.bzl", "node_base")
load("//bases/bazel/go:defs.bzl", "go_base")
load("//bases/build:defs.bzl", "build_base")

def rocky8():
    container_pull(
        name = "rocky8",
        registry = "quay.io",
        repository = "rockylinux/rockylinux",
        digest = "sha256:f0d7460b97156f6c8ea2ae73152bc11fe410d272387d60ddff36dfcea22ef689",  # 8.4
    )

def containers():
    rocky8()
    node_base()
    go_base()
    build_base()
