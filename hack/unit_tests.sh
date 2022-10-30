#!/usr/bin/env bash

set -o errexit
set -x

source hack/bazel_setup.sh

$BAZEL_T //...
