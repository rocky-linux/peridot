#!/usr/bin/env bash

set -o errexit

source hack/bazel_setup.sh

bazel_bin="$(which bazel)"
workspace_dir="$(pwd)"

$BAZEL_B //:bazel-diff

BASE_HASH="$PULL_BASE_SHA"
TARGET_HASH="$PULL_PULL_SHA"
if [[ -z "$TARGET_HASH" ]]; then
  BASE_HASH="$(git log "HEAD@{1}" --pretty=format:"%H" --merges -n 1)"
  TARGET_HASH="$PULL_BASE_SHA"
fi

# Generate starting hashes
echo "Base hash is $BASE_HASH"
git checkout "$BASE_HASH" --quiet
bazel-bin/bazel-diff generate-hashes -w "$workspace_dir" -b "$bazel_bin" starting_hashes_json

# Generate ending hashes
echo "Target hash is $TARGET_HASH"
git checkout "$TARGET_HASH" --quiet
bazel-bin/bazel-diff generate-hashes -w "$workspace_dir" -b "$bazel_bin" ending_hashes_json

# Get impacted targets
bazel-bin/bazel-diff get-impacted-targets -sh starting_hashes_json -fh ending_hashes_json -o impacted_targets


