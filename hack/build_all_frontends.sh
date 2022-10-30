#!/usr/bin/env bash

set -o errexit

source hack/bazel_setup.sh

$BAZEL_B $($BAZEL_QR "attr(tags, 'resf_frontend_bundle', //...)")
