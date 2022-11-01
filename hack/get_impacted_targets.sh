#!/usr/bin/env bash

set -o errexit

source hack/bazel_setup.sh

bazel_bin="$(which bazel)"
workspace_dir="$(pwd)"

$BAZEL_B //:bazel-diff

# Generate starting hashes
echo "Base hash is $PULL_BASE_SHA"
git checkout "$PULL_BASE_SHA" --quiet
bazel-bin/bazel-diff generate-hashes -w "$workspace_dir" -b "$bazel_bin" starting_hashes_json

# Generate ending hashes
TARGET_HASH="$PULL_PULL_SHA"
if [[ -z "$TARGET_HASH" ]]; then
  TARGET_HASH="$(git log --pretty=format:"%H" --merges -n 1)"
fi
echo "Target hash is $TARGET_HASH"
git checkout "$TARGET_HASH" --quiet
bazel-bin/bazel-diff generate-hashes -w "$workspace_dir" -b "$bazel_bin" ending_hashes_json

# Get impacted targets
bazel-bin/bazel-diff get-impacted-targets -sh starting_hashes_json -fh ending_hashes_json -o impacted_targets


