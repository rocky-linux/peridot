load("@openapi_tools_generator_bazel//:defs.bzl", "openapi_tools_generator_bazel_repositories")

def openapitools_deps():
    openapi_tools_generator_bazel_repositories(
        openapi_generator_cli_version = "5.1.0",
        sha256 = "62f9842f0fcd91e4afeafc33f19a7af41f2927c7472c601310cedfc72ff1bb19",
    )
