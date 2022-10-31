def _helm_chart_impl(ctx):
    """
    :param ctx

    Package k8s manifests into a Helm chart with specified Chart.yaml
    Sets default registry and tag to stable values defined in tools.sh / .envrc

    The following variables will be set in _helpers.tpl:
      * {name}.name
      * {name}.fullname
      * {name}.chart

    The following values.yaml variables will be available:
      * awsRegion (optional, default is us-east-2)
      * stage (required, example: dev)
      * image.repository (optional, default is STABLE_OCI_REGISTRY/STABLE_OCI_REGISTRY_REPO/image_name)
      * image.tag (optional, default is STABLE_BUILD_TAG)
      * {portName}.ingressHost (required if any ports/services are marked as exposed, port name can for example be "http" or "grpc")
      * postgresHostPort (required if service is backend or databaseUrl is not set)
      * databaseUrl (optional, required if postgresHostPort is not set)

    internal:
    The general structure of a Helm chart should be as follows:
      * Chart.yaml
      * values.yaml
      * .helmignore
      * templates/
        * _helpers.tpl
        * ctx.files.srcs

    If resfdeploy templates are used, the following values.yaml variables changes defaults:
      * image.repository (sourced from deployment.yaml)
      * image.tag (sourced from deployment.yaml)
    """

    # The stamp files should be brought in to be able to apply default registry and tag
    stamp_files = [ctx.info_file, ctx.version_file]

    # Create new staging directory
    staging_dir = "chart"
    tmp_dirname = ""

    # Declare inputs to the final Helm script
    inputs = []

    # Fail if srcs contains a file called Chart.yaml, values.yaml, _helpers.tpl or .helmignore
    for src in ctx.files.srcs:
        if src.basename in ["Chart.yaml", "values.yaml", "_helpers.tpl", ".helmignore"]:
            fail("{} is a reserved file name and should not exist in srcs".format(src.basename))

    # Copy srcs into staging directory
    for src in ctx.files.srcs + ctx.files.chart_yaml + ctx.files.values_yaml + ctx.files._helpers_tpl + ctx.files._helmignore:
        cp_out = ctx.actions.declare_file(staging_dir + "/" + src.basename)
        if tmp_dirname == "":
            tmp_dirname = cp_out.dirname
        inputs.append(cp_out)

        ctx.actions.run_shell(
            outputs = [cp_out],
            inputs = [src],
            mnemonic = "HelmCopyToStaging",
            arguments = [src.path, cp_out.path],
            command = "cp -RL $1 $2",
        )

    # Expand template for Helm script
    tarball_file = ctx.actions.declare_file(ctx.label.name + ".tgz")
    out_file = ctx.actions.declare_file(ctx.label.name + ".helm.bash")
    ctx.actions.expand_template(
        template = ctx.file._helm_script,
        output = out_file,
        substitutions = {
            "TMPL_helm_bin": ctx.file._helm_bin.path,
            "TMPL_yq_bin": ctx.file._yq_bin.path,
            "TMPL_name": ctx.attr.package_name,
            "TMPL_staging_dir": tmp_dirname,
            "TMPL_image_name": ctx.attr.package_name if not ctx.attr.image_name else ctx.attr.image_name,
            "TMPL_tarball_file_path": tarball_file.path,
            "TMPL_stamp_files": ";".join([x.path for x in stamp_files]),
        },
        is_executable = True,
    )

    # Run Helm script and generate a tarball
    ctx.actions.run(
        outputs = [tarball_file],
        inputs = inputs + stamp_files + [ctx.file._helm_bin, ctx.file._yq_bin],
        executable = out_file,
        mnemonic = "HelmChart",
    )

    return [DefaultInfo(
        files = depset([tarball_file]),
    )]

helm_chart = rule(
    implementation = _helm_chart_impl,
    attrs = {
        "package_name": attr.string(
            doc = "The name of the package",
            mandatory = True,
        ),
        "image_name": attr.string(
            doc = "The name of the OCI image, defaults to package_name. Ignored if resfdeploy is used and sourced from deployment.yaml",
        ),
        "chart_yaml": attr.label(
            doc = "Chart.yaml file path",
            default = ":Chart.yaml",
            mandatory = True,
            allow_single_file = True,
        ),
        "values_yaml": attr.label(
            doc = "values.yaml file path",
            default = ":values.yaml",
            mandatory = True,
            allow_single_file = True,
        ),
        "srcs": attr.label_list(
            doc = "List of templates/manifests to be included in chart",
            mandatory = True,
            allow_files = True,
        ),
        "_helm_bin": attr.label(
            doc = "Helm binary path",
            default = "//rules_resf/internal/helm:helm_tool",
            allow_single_file = True,
            cfg = "host",
        ),
        "_yq_bin": attr.label(
            doc = "yq binary path",
            default = "//rules_resf/internal/helm:yq_tool",
            allow_single_file = True,
            cfg = "host",
        ),
        "_helpers_tpl": attr.label(
            doc = "Helpers template path",
            default = "//rules_resf/internal/helm:_helpers.tpl",
            allow_single_file = True,
        ),
        "_helm_script": attr.label(
            doc = "Helm script path",
            default = "//rules_resf/internal/helm:helm.bash",
            allow_single_file = True,
        ),
        "_helmignore": attr.label(
            doc = "Helmignore file path",
            default = "//rules_resf/internal/helm:.helmignore",
            allow_single_file = True,
        ),
    },
)
