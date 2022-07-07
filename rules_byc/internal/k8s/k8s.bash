#!/usr/bin/env bash

#
# Copyright (c) All respective contributors to the Peridot Project. All rights reserved.
# Copyright (c) 2021-2022 Rocky Enterprise Software Foundation, Inc. All rights reserved.
# Copyright (c) 2021-2022 Ctrl IQ, Inc. All rights reserved.
#
# Redistribution and use in source and binary forms, with or without
# modification, are permitted provided that the following conditions are met:
#
# 1. Redistributions of source code must retain the above copyright notice,
# this list of conditions and the following disclaimer.
#
# 2. Redistributions in binary form must reproduce the above copyright notice,
# this list of conditions and the following disclaimer in the documentation
# and/or other materials provided with the distribution.
#
# 3. Neither the name of the copyright holder nor the names of its contributors
# may be used to endorse or promote products derived from this software without
# specific prior written permission.
#
# THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
# AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
# IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
# ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
# LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
# CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
# SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
# INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
# CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
# ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
# POSSIBILITY OF SUCH DAMAGE.
#

# --- begin runfiles.bash initialization ---
# Copy-pasted from Bazel's Bash runfiles library (tools/bash/runfiles/runfiles.bash).
set -o pipefail
if [[ ! -d "${RUNFILES_DIR:-/dev/null}" && ! -f "${RUNFILES_MANIFEST_FILE:-/dev/null}" ]]; then
  if [[ -f "$0.runfiles_manifest" ]]; then
    export RUNFILES_MANIFEST_FILE="$0.runfiles_manifest"
  elif [[ -f "$0.runfiles/MANIFEST" ]]; then
    export RUNFILES_MANIFEST_FILE="$0.runfiles/MANIFEST"
  elif [[ -f "$0.runfiles/bazel_tools/tools/bash/runfiles/runfiles.bash" ]]; then
    export RUNFILES_DIR="$0.runfiles"
  fi
fi
if [[ -f "${RUNFILES_DIR:-/dev/null}/bazel_tools/tools/bash/runfiles/runfiles.bash" ]]; then
  source "${RUNFILES_DIR}/bazel_tools/tools/bash/runfiles/runfiles.bash"
elif [[ -f "${RUNFILES_MANIFEST_FILE:-/dev/null}" ]]; then
  source "$(grep -m1 "^bazel_tools/tools/bash/runfiles/runfiles.bash " \
            "$RUNFILES_MANIFEST_FILE" | cut -d ' ' -f 2-)"
else
  echo >&2 "ERROR: cannot find @bazel_tools//tools/bash/runfiles:runfiles.bash"
  exit 1
fi
# --- end runfiles.bash initialization ---
# Export RUNFILES_* envvars (and a couple more) for subprocesses.
runfiles_export_envvars

IFS=';' read -ra KCTL <<< "TMPL_files"
IFS=';' read -ra KCWA <<< "TMPL_wait"
IFS=';' read -ra KCRE <<< "TMPL_replace"

KAPPLY="kubectl apply"

for i in "${KCTL[@]}"; do
  if [[ "$i" == *"deployment"* ]]; then
    if [[ "$STABLE_REGISTRY_SECRET" != "none" ]]; then
      COPY_TO_NS=$(echo "{$(cat ${i} | grep "namespace" | head -n 1)}" | jq -r '.namespace' | tr -d '\n')
      kubectl -n "registry-secret${STABLE_STAGE}" get secret registry -o json | jq ".metadata.namespace=\"${COPY_TO_NS}\"" | kubectl apply --force -f -
    fi
  fi

  if [[ "${KCRE[*]}" =~ ${i} ]]; then
    kubectl replace --force -f "${i}"
  else
    if [[ "$STABLE_STAGE" == "-dev" && "${i}" == *"deployment"* ]]; then
      KAPPLY="kubectl replace"
    fi
    $KAPPLY --force -f "${i}"
  fi

  if [[ "${KCWA[*]}" =~ ${i} ]]; then
    kubectl wait --for=condition=complete --timeout=190s \
      -f "${i}"
  fi
done


