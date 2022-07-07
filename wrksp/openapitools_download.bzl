load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def openapitools_download():
    http_archive(
        name = "openapi_tools_generator_bazel",
        sha256 = "c6e4c253f1ae0fbe4d4ded8a719f6647273141d0dc3c0cd8bb074aa7fc3c8d1c",
        urls = ["https://github.com/OpenAPITools/openapi-generator-bazel/releases/download/0.1.5/openapi-tools-generator-bazel-0.1.5.tar.gz"],
    )
