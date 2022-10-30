load("@bazel_skylib//lib:shell.bzl", "shell")

def _k8s_apply_impl(ctx):
    runfiles = ctx.runfiles().merge(ctx.attr._bash_runfiles[DefaultInfo].default_runfiles)
    runfiles = runfiles.merge(ctx.runfiles(
        files = ctx.files.srcs,
    ))
    files = [y for y in [x[DefaultInfo].files.to_list() for x in ctx.attr.srcs]]
    all_files = [x for y in files for x in y]
    sts = ";".join([y.short_path for y in all_files])

    wait_sts = ";".join([y.short_path for y in ctx.files.wait])
    replace_sts = ";".join([y.short_path for y in ctx.files.replace])

    out_file = ctx.actions.declare_file(ctx.label.name + ".bash")
    ctx.actions.expand_template(
        template = ctx.file._k8s_bash,
        output = out_file,
        substitutions = {
            "TMPL_files": sts,
            "TMPL_wait": wait_sts,
            "TMPL_replace": replace_sts,
        },
        is_executable = True,
    )
    return [DefaultInfo(
        files = depset([out_file]),
        runfiles = runfiles,
        executable = out_file,
    )]

k8s_apply = rule(
    implementation = _k8s_apply_impl,
    attrs = {
        # Kubernetes json or yaml files
        # Can be a target that generates them
        # such as jsonnet_to_json
        "srcs": attr.label_list(
            allow_files = True,
            mandatory = True,
        ),
        # Files to wait for
        # Ex. jobs such as migrations
        "wait": attr.label_list(
            allow_files = True,
        ),
        # Resources to completely replace
        # Ex. to clean up hanging jobs
        "replace": attr.label_list(
            allow_files = True,
        ),
        # Namespace to be used in delete
        "namespace": attr.string(),
        # Used if resource should be deleted
        "delete": attr.bool(
            default = False,
        ),
        "_bash_runfiles": attr.label(
            default = Label("@bazel_tools//tools/bash/runfiles"),
        ),
        "_k8s_bash": attr.label(
            default = Label("//rules_resf/internal/k8s:k8s.bash"),
            allow_single_file = True,
        ),
    },
    executable = True,
)
