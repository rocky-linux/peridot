#!/usr/bin/env bash

set -o errexit

source hack/bazel_setup.sh

$BAZEL_T //...
