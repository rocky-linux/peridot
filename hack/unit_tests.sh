#!/usr/bin/env bash

set -o errexit
set -x

bazel test --config=ci //...
