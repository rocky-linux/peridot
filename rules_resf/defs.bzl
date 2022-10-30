load("//rules_resf/internal/resf_bundle:resf_bundle.bzl", _resf_bundle = "resf_bundle", _resf_bundle_run = "resf_bundle_run")
load("//rules_resf/internal/k8s:k8s.bzl", _k8s_apply = "k8s_apply")
load("//rules_resf/internal/container:container.bzl", _container = "container", _migration_tar = "migration_tar")
load("@io_bazel_rules_jsonnet//jsonnet:jsonnet.bzl", "jsonnet_to_json")
load("@build_bazel_rules_nodejs//:index.bzl", "nodejs_binary")
load("@com_github_atlassian_bazel_tools//:multirun/def.bzl", "multirun")

resf_bundle = _resf_bundle
k8s_apply = _k8s_apply
container = _container
migration_tar = _migration_tar

RESFDEPLOY_OUTS_BASE = [
    "001-ns-sa.yaml",
    "003-deployment.yaml",
    "004-svc-vs-dr.yaml",
]

RESFDEPLOY_OUTS_MIGRATE = RESFDEPLOY_OUTS_BASE + [
    "002-migrate.yaml",
]

RESFDEPLOY_OUTS_CUSTOM = RESFDEPLOY_OUTS_BASE + [
    "005-custom.yaml",
]

RESFDEPLOY_OUTS_MIGRATE_CUSTOM = RESFDEPLOY_OUTS_BASE + [
    "002-migrate.yaml",
    "005-custom.yaml",
]

def tag_default_update(defaults, append):
    tdict = defaults
    tdict.update(append)
    return tdict

# to find the correct kind during ci run
def peridot_k8s(name, src, tags = [], outs = [], static = False, prod_only = False, dependent_push = [], force_normal_tags = False, **kwargs):
    ext_str_nested = "{STABLE_OCI_REGISTRY_NO_NESTED_SUPPORT_IN_2022_SHAME_ON_YOU_AWS}"
    if force_normal_tags:
        ext_str_nested = "false"
    ext_strs = {
        "tag": "{STABLE_BUILD_TAG}",
        "stage": "{STABLE_STAGE}",
        "local_environment": "{STABLE_LOCAL_ENVIRONMENT}",
        "user": "{BUILD_USER}",
        "oci_registry": "{STABLE_OCI_REGISTRY}",
        "oci_registry_repo": "{STABLE_OCI_REGISTRY_REPO}",
        "oci_registry_docker": "{STABLE_OCI_REGISTRY_DOCKER}",
        "oci_registry_no_nested_support_in_2022_shame_on_you_aws": ext_str_nested,
        "domain_user": "{STABLE_DOMAIN_USER}",
        "registry_secret": "{STABLE_REGISTRY_SECRET}",
        "site": "{STABLE_SITE}",
    }
    jsonnet_to_json(
        name = name,
        src = src,
        outs = outs,
        tags = tags + [
            "manual",
            "peridot_k8s",
        ],
        ext_strs = select({
            "//platforms:arm64": dict(ext_strs, arch = "arm64"),
            "//platforms:x86_64": dict(ext_strs, arch = "amd64"),
            "//platforms:s390x": dict(ext_strs, arch = "s390x"),
            "//platforms:ppc64le": dict(ext_strs, arch = "ppc64le"),
        }),
        stamp_keys = [
            "tag",
            "stage",
            "local_environment",
            "user",
            "oci_registry",
            "oci_registry_repo",
            "oci_registry_docker",
            "oci_registry_no_nested_support_in_2022_shame_on_you_aws",
            "domain_user",
            "registry_secret",
            "site",
        ],
        multiple_outputs = True,
        extra_args = ["-S"],
        **kwargs
    )

    k8s_apply(
        name = "%s.apply" % name,
        srcs = [":%s" % name],
        tags = ["manual"],
        visibility = ["//visibility:public"],
    )
    multirun(
        name = "%s.push" % name,
        commands = dependent_push + [":%s_container" % name],
        tags = ["manual"],
    )
    multirun(
        name = "%s.push_apply" % name,
        commands = [
            ":%s.push" % name,
            ":%s.apply" % name,
        ],
        tags = ["manual"],
    )

def resf_frontend(name, tags = [], **kwargs):
    _resf_bundle(
        name = "{}.bundle".format(name),
        build = True,
        tags = tags + [
            "manual",
            "resf_frontend_bundle",
        ],
        **kwargs
    )

    _resf_bundle_run(
        name = "{}.server".format(name),
        build = False,
        tags = tags + [
            "manual",
            "resf_frontend_server",
            "ibazel_notify_changes",
            "ibazel_live_reload",
        ],
        **kwargs
    )
