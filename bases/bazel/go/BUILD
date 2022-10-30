load("@io_bazel_rules_docker//container:container.bzl", "container_image")

container_image(
    name = "go",
    base = select({
        "//platforms:arm64": "@go_base_arm64//image",
        "//platforms:x86_64": "@go_base_amd64//image",
        "//platforms:s390x": "@go_base_s390x//image",
        "//platforms:ppc64le": "@go_base_ppc64le//image",
    }),
    tags = ["manual"],
    visibility = ["//visibility:public"],
)

container_image(
    name = "go_arm64",
    base = "@go_base_arm64//image",
    tags = ["manual"],
    visibility = ["//visibility:public"],
)

container_image(
    name = "go_amd64",
    base = "@go_base_amd64//image",
    tags = ["manual"],
    visibility = ["//visibility:public"],
)

container_image(
    name = "go_s390x",
    base = "@go_base_s390x//image",
    tags = ["manual"],
    visibility = ["//visibility:public"],
)

container_image(
    name = "go_ppc64le",
    base = "@go_base_ppc64le//image",
    tags = ["manual"],
    visibility = ["//visibility:public"],
)
