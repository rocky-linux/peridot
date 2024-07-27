load("@rules_pkg//:pkg.bzl", "pkg_tar")
load("@io_bazel_rules_docker//container:container.bzl", "container_image", "container_layer", "container_push")
load("@io_bazel_rules_docker//nodejs:image.bzl", "nodejs_image")

REGISTRY_VARIANT = "aws"

def migration_tar():
    pkg_tar(
        name = "migrate",
        srcs = native.glob(["*.sql"]),
        package_dir = "/migrations",
    )

def container(image_name, files, tars_to_layer = [], base = "//bases/bazel/go", registry = "{STABLE_OCI_REGISTRY}", repository = "{STABLE_OCI_REGISTRY_REPO}", full_img_path = None, frontend = False, server_files = [], server_entrypoint = None, architecture = None, force_normal_tags = False, disable_conditional = False):
    container_layer(
        name = "%s_bin" % image_name,
        directory = "/home/app/%s" % "bundle" if frontend else "bundle",
        files = files,
        tags = ["manual"],
        visibility = [":__subpackages__"],
    )

    extra_layers = []
    if len(tars_to_layer) > 0:
        layer_name = "%s_tar_layer" % image_name
        container_layer(
            name = layer_name,
            tags = ["manual"],
            tars = tars_to_layer,
            visibility = [":__subpackages__"],
        )
        extra_layers.append(layer_name)

    if not architecture:
        container_image(
            name = "%s_image" % image_name,
            architecture = select({
                "//platforms:arm64": "arm64",
                "//platforms:x86_64": "amd64",
                "//platforms:s390x": "s390x",
                "//platforms:ppc64le": "ppc64le",
            }),
            base = base,
            layers = [":%s_bin" % image_name] + extra_layers,
            tags = ["manual"],
            visibility = ["//visibility:public"],
        )
    else:
        container_image(
            name = "%s_image" % image_name,
            architecture = architecture,
            base = base,
            layers = [":%s_bin" % image_name] + extra_layers,
            tags = ["manual"],
            visibility = ["//visibility:public"],
        )

    tag = "{STABLE_BUILD_TAG}"
    img_path = "%s/%s" % (repository, image_name)
    if full_img_path != None:
        img_path = full_img_path

    should_use_aws_format = full_img_path == None and REGISTRY_VARIANT == "aws" and not force_normal_tags
    if should_use_aws_format:
        tag = "%s-{STABLE_BUILD_TAG}" % image_name
        img_path = repository

    if len(server_files) > 0:
        nodejs_image(
            name = "%s_image_node" % image_name,
            entry_point = server_entrypoint,
            data = server_files,
            base = ":%s_image" % image_name,
            tags = ["manual"],
        )

    container_push(
        name = "%s_container" % image_name,
        format = "Docker",
        image = (":%s_image_node" if server_entrypoint else ":%s_image") % image_name,
        registry = registry,
        repository = select({
            "//platforms:arm64": img_path,
            "//platforms:x86_64": img_path,
            "//platforms:s390x": "%s/%s_s390x" % (repository, image_name),
            "//platforms:ppc64le": "%s/%s_ppc64le" % (repository, image_name),
        }) if not should_use_aws_format and not disable_conditional else img_path,
        tag = select({
            "//platforms:arm64": tag,
            "//platforms:x86_64": tag,
            "//platforms:s390x": "%s_s390x-{STABLE_BUILD_TAG}" % image_name,
            "//platforms:ppc64le": "%s_ppc64le-{STABLE_BUILD_TAG}" % image_name,
        }) if should_use_aws_format and not disable_conditional else tag,
        tags = ["manual"],
        visibility = ["//visibility:public"],
    )
