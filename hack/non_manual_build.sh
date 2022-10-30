#!/usr/bin/env bash

set -o errexit
set -x

bazel build --config=ci $(bazel query "//... except attr(tags, 'manual', //...) except //vendor/...")
