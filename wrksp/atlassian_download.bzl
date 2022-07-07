load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")
ATLASSIAN_COMMIT_HASH = "6fbc36c639a8f376182bb0057dd557eb2440d4ed"
def atlassian_download():
    http_archive(
        name = "com_github_atlassian_bazel_tools",
        sha256 = "6b438f4d8c698f69ed4473cba12da3c3a7febf90ce8e3c383533d5a64d8c8f19",
        strip_prefix = "bazel-tools-%s" % ATLASSIAN_COMMIT_HASH,
        urls = ["https://github.com/atlassian/bazel-tools/archive/%s.tar.gz" % ATLASSIAN_COMMIT_HASH],
    )
