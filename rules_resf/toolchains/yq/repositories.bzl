load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

build_file_content = """
exports_files(["yq"])
"""

patch_cmds = ["mv yq_* yq"]

def yq_repositories():
    http_archive(
        name = "yq_linux_x86_64",
        sha256 = "29716620085fdc7e3d2d12a749124a5113091183306a274f8abc61009ca38996",
        urls = ["https://github.com/mikefarah/yq/releases/download/v4.25.2/yq_linux_amd64.tar.gz"],
        patch_cmds = patch_cmds,
        build_file_content = build_file_content,
    )

    http_archive(
        name = "yq_linux_arm64",
        sha256 = "77d84462f65c4f4d9a972158887dcd35c029cf199ee9c42b573a6e6e6ecd372f",
        urls = ["https://github.com/mikefarah/yq/releases/download/v4.25.2/yq_linux_arm64.tar.gz"],
        patch_cmds = patch_cmds,
        build_file_content = build_file_content,
    )

    http_archive(
        name = "yq_darwin_x86_64",
        sha256 = "b7a836729142a6f54952e9a7675ae183acb7fbacc36ff555ef763939a26731a6",
        urls = ["https://github.com/mikefarah/yq/releases/download/v4.25.2/yq_darwin_amd64.tar.gz"],
        patch_cmds = patch_cmds,
        build_file_content = build_file_content,
    )

    http_archive(
        name = "yq_darwin_arm64",
        sha256 = "e2e8fe89ee4d4e7257838e5941c50ef5aa753a86c699ade8a099cd46f09da1d3",
        urls = ["https://github.com/mikefarah/yq/releases/download/v4.25.2/yq_darwin_arm64.tar.gz"],
        patch_cmds = patch_cmds,
        build_file_content = build_file_content,
    )
