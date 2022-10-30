#!/usr/bin/env bash

set -o errexit

source hack/bazel_setup.sh

bazel_bin="$(which bazel)"
workspace_dir="$(pwd)"

$BAZEL_B //:bazel-diff

# Generate starting hashes
git checkout "$PULL_BASE_SHA" --quiet
bazel-bin/bazel-diff generate-hashes -w "$workspace_dir" -b "$bazel_bin" starting_hashes_json

# Generate ending hashes
git checkout "$PULL_PULL_SHA" --quiet
bazel-bin/bazel-diff generate-hashes -w "$workspace_dir" -b "$bazel_bin" ending_hashes_json

# Get impacted targets
bazel-bin/bazel-diff get-impacted-targets -sh starting_hashes_json -fh ending_hashes_json impacted_targets

# Build impacted targets
hack/build_impacted_frontend.sh


