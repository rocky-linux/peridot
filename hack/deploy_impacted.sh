#!/usr/bin/env bash

set -o errexit

source .envrc.prod.resf
source hack/bazel_setup.sh

hack/get_impacted_targets.sh

return_if_impacted_targets_empty

aws eks --region us-east-2 update-kubeconfig --name peridot-T8WbrA

AWS_JWT="$(aws ecr get-login-password --region us-east-2)"
B64_AWS_AUTH="$(echo -n "AWS:$AWS_JWT" | base64 -w 0)"
mkdir -p ~/.docker
echo '{"auths":{"893168113496.dkr.ecr.us-east-2.amazonaws.com":{"auth":"'"$B64_AWS_AUTH"'"}}}' > ~/.docker/config.json

starting_query="attr(tags, 'push_apply',"

for t in `cat impacted_targets`; do
  starting_query="$starting_query $t union"
done

starting_query=${starting_query%" union"}
starting_query="$starting_query)"

TARGETS=$($BAZEL_QR "$starting_query")
for target in $TARGETS; do
  $BAZEL_R "$target"
done


