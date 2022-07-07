load("@rules_python//python:pip.bzl", "pip_repositories")

def python_deps():
    pip_repositories()
