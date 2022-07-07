load("@io_bazel_rules_docker//container:container.bzl", "container_image")

container_image(
    name = "build_arm64",
    base = "@build_base_arm64//image",
    visibility = ["//visibility:public"],
)

container_image(
    name = "build_amd64",
    base = "@build_base_amd64//image",
    visibility = ["//visibility:public"],
)

container_image(
    name = "build_s390x",
    base = "@build_base_s390x//image",
    visibility = ["//visibility:public"],
)

container_image(
    name = "build_ppc64le",
    base = "@build_base_ppc64le//image",
    visibility = ["//visibility:public"],
)

container_image(
    name = "build",
    base = select({
        "//platforms:arm64": "@build_base_arm64//image",
        "//platforms:x86_64": "@build_base_amd64//image",
        "//platforms:s390x": "@build_base_s390x//image",
        "//platforms:ppc64le": "@build_base_ppc64le//image",
    }),
    visibility = ["//visibility:public"],
)
