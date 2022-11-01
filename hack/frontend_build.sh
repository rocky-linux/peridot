#!/usr/bin/env bash

set -o errexit

source hack/bazel_setup.sh

hack/get_impacted_targets.sh

return_if_impacted_targets_empty

# Build impacted targets
hack/build_impacted_frontend.sh
