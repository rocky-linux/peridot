#!/usr/bin/env bash

export SILO_KEY=rocky86-peridot-prow-1
export REMOTE_DEF=cache-silo-key=$SILO_KEY

CI=${CI:-}
QUERY_FLAGS=""
FLAGS="--show_progress_rate_limit=5 --color=yes --ui_actions_shown=30 --terminal_columns=140 --show_timestamps --verbose_failures --announce_rc --experimental_repository_cache_hardlinks --disk_cache= --sandbox_tmpfs_path=/tmp --experimental_guard_against_concurrent_changes"

if [[ -n ${CI} ]]; then
  FLAGS="--config=ci $FLAGS"
  QUERY_FLAGS="--config=ci"
fi

BAZEL_B="bazel $SOPTIONS build $FLAGS --remote_default_exec_properties=$REMOTE_DEF"
BAZEL_R="bazel $SOPTIONS run $FLAGS --remote_default_exec_properties=$REMOTE_DEF"
BAZEL_T="bazel $SOPTIONS test $FLAGS --remote_default_exec_properties=$REMOTE_DEF --test_arg=-test.v --flaky_test_attempts=3 --build_tests_only --test_output=errors"
BAZEL_QR="bazel $SOPTIONS query $QUERY_FLAGS --keep_going --noshow_progress"

return_if_impacted_targets_empty() {
  if [[ -z "$(cat impacted_targets)" ]]; then
    echo "No impacted targets found"
    exit 0
  fi
}
