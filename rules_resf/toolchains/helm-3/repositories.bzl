load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

build_file_content = """
exports_files(["helm"])
"""

patch_cmds = ["mv */helm helm"]

def helm3_repositories():
    http_archive(
        name = "helm3_linux_x86_64",
        sha256 = "1484ffb0c7a608d8069470f48b88d729e88c41a1b6602f145231e8ea7b43b50a",
        urls = ["https://get.helm.sh/helm-v3.9.0-linux-amd64.tar.gz"],
        patch_cmds = patch_cmds,
        build_file_content = build_file_content,
    )

    http_archive(
        name = "helm3_linux_arm64",
        sha256 = "5c0aa709c5aaeedd190907d70f9012052c1eea7dff94bffe941b879a33873947",
        urls = ["https://get.helm.sh/helm-v3.9.0-linux-arm64.tar.gz"],
        patch_cmds = patch_cmds,
        build_file_content = build_file_content,
    )

    http_archive(
        name = "helm3_darwin_x86_64",
        sha256 = "7e5a2f2a6696acf278ea17401ade5c35430e2caa57f67d4aa99c607edcc08f5e",
        urls = ["https://get.helm.sh/helm-v3.9.0-darwin-amd64.tar.gz"],
        patch_cmds = patch_cmds,
        build_file_content = build_file_content,
    )

    http_archive(
        name = "helm3_darwin_arm64",
        sha256 = "22cf080ded5dd71ec15d33c13586ace9b6002e97518a76df628e67ecedd5aa70",
        urls = ["https://get.helm.sh/helm-v3.9.0-darwin-arm64.tar.gz"],
        patch_cmds = patch_cmds,
        build_file_content = build_file_content,
    )
