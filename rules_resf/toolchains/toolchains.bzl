load("//rules_resf/toolchains/helm-3:repositories.bzl", "helm3_repositories")
load("//rules_resf/toolchains/yq:repositories.bzl", "yq_repositories")

def toolchains_repositories():
    helm3_repositories()
    yq_repositories()
