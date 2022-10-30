#!/usr/bin/env bash

set -o errexit

source hack/bazel_setup.sh

starting_query="attr(tags, 'resf_frontend_bundle',"

for t in `cat impacted_targets`; do
  starting_query="$starting_query $t union"
done

starting_query=${starting_query%" union"}
starting_query="$starting_query)"

$BAZEL_B $($BAZEL_QR "$starting_query")
