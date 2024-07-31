workspace(
    name = "peridot",
    managed_directories = {"@npm": ["node_modules"]},
)

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive", "http_jar")

http_archive(
    name = "bazel_skylib",
    sha256 = "74d544d96f4a5bb630d465ca8bbcfe231e3594e5aae57e1edbf17a6eb3ca2506",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/bazel-skylib/releases/download/1.3.0/bazel-skylib-1.3.0.tar.gz",
        "https://github.com/bazelbuild/bazel-skylib/releases/download/1.3.0/bazel-skylib-1.3.0.tar.gz",
    ],
)

# --start python--
load("//wrksp:python_download.bzl", "python_download")

python_download()

load("//wrksp:python_deps.bzl", "python_deps")

python_deps()
# --end python--

http_archive(
    name = "com_google_protobuf",
    sha256 = "d19643d265b978383352b3143f04c0641eea75a75235c111cc01a1350173180e",
    strip_prefix = "protobuf-25.3",
    urls = ["https://github.com/protocolbuffers/protobuf/archive/v25.3.tar.gz"],
)

load("@com_google_protobuf//:protobuf_deps.bzl", "protobuf_deps")

protobuf_deps()

http_archive(
    name = "io_bazel_rules_go",
    patch_args = ["-p1"],
    patches = [
        "//patches:0001-Disable-ppc64-support-in-rules_go-as-it-breaks-match.patch",
    ],
    sha256 = "af47f30e9cbd70ae34e49866e201b3f77069abb111183f2c0297e7e74ba6bbc0",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/rules_go/releases/download/v0.47.0/rules_go-v0.47.0.zip",
        "https://github.com/bazelbuild/rules_go/releases/download/v0.47.0/rules_go-v0.47.0.zip",
    ],
)

http_archive(
    name = "bazel_gazelle",
    integrity = "sha256-MpOL2hbmcABjA1R5Bj2dJMYO2o15/Uc5Vj9Q0zHLMgk=",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/bazel-gazelle/releases/download/v0.35.0/bazel-gazelle-v0.35.0.tar.gz",
        "https://github.com/bazelbuild/bazel-gazelle/releases/download/v0.35.0/bazel-gazelle-v0.35.0.tar.gz",
    ],
)

load("@io_bazel_rules_go//go:deps.bzl", "go_register_toolchains", "go_rules_dependencies")
load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies", "go_repository")
load("//:repositories.bzl", "go_repositories")

go_rules_dependencies()

http_archive(
    name = "rules_pkg",
    sha256 = "d250924a2ecc5176808fc4c25d5cf5e9e79e6346d79d5ab1c493e289e722d1d0",
    urls = [
        "https://github.com/bazelbuild/rules_pkg/releases/download/0.10.1/rules_pkg-0.10.1.tar.gz",
    ],
)

load("@rules_pkg//:deps.bzl", "rules_pkg_dependencies")

rules_pkg_dependencies()

load("@rules_pkg//toolchains/rpm:rpmbuild_configure.bzl", "find_system_rpmbuild")

find_system_rpmbuild(name = "rules_pkg_rpmbuild")

go_register_toolchains(
    nogo = "@peridot//:nogo",
    version = "1.22.2",
)

# gazelle:repository_macro repositories.bzl%go_repositories
go_repositories()

go_repository(
    name = "org_golang_google_grpc",
    build_file_generation = "on",
    build_file_proto_mode = "disable",
    importpath = "google.golang.org/grpc",
    sum = "h1:HQKZ/fa1bXkX1oFOvSjmZEUL8wLSaZTjCcLAlmZRtdk=",
    version = "v1.62.0",
)

go_repository(
    name = "org_golang_x_net",
    importpath = "golang.org/x/net",
    sum = "h1:AQyQV4dYCvJ7vGmJyKki9+PBdyvhkSd8EIx/qb0AYv4=",
    version = "v0.21.0",
)

go_repository(
    name = "org_golang_x_text",
    importpath = "golang.org/x/text",
    sum = "h1:ScX5w1eTa3QqT8oi6+ziP7dTV1S2+ALU0bI+0zXKWiQ=",
    version = "v0.14.0",
)

go_repository(
    name = "org_golang_x_oauth2",
    importpath = "golang.org/x/oauth2",
    sum = "h1:6m3ZPmLEFdVxKKWnKq4VqZ60gutO35zm+zrAHVmHyDQ=",
    version = "v0.17.0",
)

gazelle_dependencies()

http_archive(
    name = "googleapis",
    sha256 = "9d1a930e767c93c825398b8f8692eca3fe353b9aaadedfbcf1fca2282c85df88",
    strip_prefix = "googleapis-64926d52febbf298cb82a8f472ade4a3969ba922",
    urls = [
        "https://github.com/googleapis/googleapis/archive/64926d52febbf298cb82a8f472ade4a3969ba922.zip",
    ],
)

load("@googleapis//:repository_rules.bzl", "switched_rules_by_language")

switched_rules_by_language(
    name = "com_google_googleapis_imports",
    go = True,
    grpc = True,
)

http_archive(
    name = "build_bazel_rules_nodejs",
    sha256 = "8f5f192ba02319254aaf2cdcca00ec12eaafeb979a80a1e946773c520ae0a2c9",
    urls = ["https://github.com/bazelbuild/rules_nodejs/releases/download/3.7.0/rules_nodejs-3.7.0.tar.gz"],
)

load("@build_bazel_rules_nodejs//:index.bzl", "node_repositories", "yarn_install")

node_repositories(
    node_version = "16.2.0",
    package_json = ["//:package.json"],
    yarn_version = "1.22.10",
)

yarn_install(
    name = "npm",
    package_json = "//:package.json",
    yarn_lock = "//:yarn.lock",
)

# --start docker--
load("//wrksp:docker_download.bzl", "docker_download")

docker_download()

load("//wrksp:docker_deps.bzl", "docker_deps")

docker_deps()
# --end docker--

# --start openapitools--
load("//wrksp:openapitools_download.bzl", "openapitools_download")

openapitools_download()

load("//wrksp:openapitools_deps.bzl", "openapitools_deps")

openapitools_deps()
# --end openapitools--

# --start protoc_gen_validate--
http_archive(
    name = "com_envoyproxy_protoc_gen_validate",
    sha256 = "92e29c2150675ce954c965bcaa559ca944704b75711533cfe03ce541dcf5a1dd",
    strip_prefix = "protoc-gen-validate-1.0.4",
    urls = [
        "https://github.com/bufbuild/protoc-gen-validate/archive/v1.0.4.tar.gz",
    ],
)

load("@com_envoyproxy_protoc_gen_validate//:dependencies.bzl", "go_third_party")

go_third_party()
# --end protoc_gen_validate--

# --start jsonnet--
http_archive(
    name = "io_bazel_rules_jsonnet",
    sha256 = "fa791a38167a198a8b42bfc750ee5642f811ab20649c5517e12719e78d9a133f",
    strip_prefix = "rules_jsonnet-bd79290c53329db8bc8e3c5b709fbf822d865046",
    urls = ["https://github.com/bazelbuild/rules_jsonnet/archive/bd79290c53329db8bc8e3c5b709fbf822d865046.tar.gz"],
)

load("@io_bazel_rules_jsonnet//jsonnet:jsonnet.bzl", "jsonnet_repositories")

jsonnet_repositories()

load("@google_jsonnet_go//bazel:repositories.bzl", "jsonnet_go_repositories")

jsonnet_go_repositories()

http_archive(
    name = "cpp_jsonnet",
    sha256 = "cbbdddc82c0090881aeff0334b6d60578a15b6fafdb0ac54974840d2b7167d88",
    strip_prefix = "jsonnet-60bcf7f097ce7ec2d40ea2b4d0217d1e325f4769",
    urls = ["https://github.com/google/jsonnet/archive/60bcf7f097ce7ec2d40ea2b4d0217d1e325f4769.tar.gz"],
)
# --end jsonnet--

# --start atlassian--
load("//wrksp:atlassian_download.bzl", "atlassian_download")

atlassian_download()

load("//wrksp:atlassian_deps.bzl", "atlassian_deps")

atlassian_deps()
# --end atlassian--

# --start bazel-diff--

http_jar(
    name = "bazel_diff",
    sha256 = "a88c267227a770b787ec939b64cca907efa6e1a1c0d5c55283d7332ddb05d3b5",
    urls = [
        "https://github.com/Tinder/bazel-diff/releases/download/4.8.2/bazel-diff_deploy.jar",
    ],
)
# --end bazel-diff--

new_local_repository(
    name = "raw_ts_library",
    build_file = "//rules_raw_ts_library:BUILD",
    path = "rules_raw_ts_library",
)

load("//bases/bazel:containers.bzl", "containers")

containers()

load("//rules_resf/toolchains:toolchains.bzl", "toolchains_repositories")

toolchains_repositories()
