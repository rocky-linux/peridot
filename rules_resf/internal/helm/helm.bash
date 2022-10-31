#!/usr/bin/env bash
set -e
shopt -s extglob

HELM_BIN="$(pwd)/TMPL_helm_bin"
YQ_BIN="$(pwd)/TMPL_yq_bin"

NAME="TMPL_name"
staging_dir="TMPL_staging_dir"
image_name="TMPL_image_name"
tarball_file_path="$(pwd)/TMPL_tarball_file_path"
IFS=';' read -ra stamp_files <<< "TMPL_stamp_files"

# Find STABLE_BUILD_TAG, STABLE_OCI_REGISTRY, STABLE_OCI_REGISTRY_REPO, STABLE_OCI_REGISTRY_NO_NESTED_SUPPORT_IN_2022_SHAME_ON_YOU_AWS and make it available to the script.
vars=("STABLE_BUILD_TAG" "STABLE_OCI_REGISTRY" "STABLE_OCI_REGISTRY_REPO" "STABLE_OCI_REGISTRY_NO_NESTED_SUPPORT_IN_2022_SHAME_ON_YOU_AWS")
for stamp in "${stamp_files[@]}"; do
  for var in "${vars[@]}"; do
    if grep -q "${var} " "${stamp}"; then
      export "${var}"="$(grep "${var} " "${stamp}" | cut -d ' ' -f 2 | tr -d '\n')"
    fi
  done
done

helm_repo="${STABLE_OCI_REGISTRY}/${STABLE_OCI_REGISTRY_REPO}/$image_name"
helm_tag="${STABLE_BUILD_TAG}"
if [[ "${STABLE_OCI_REGISTRY_NO_NESTED_SUPPORT_IN_2022_SHAME_ON_YOU_AWS}" == "true" ]]; then
  helm_repo="${STABLE_OCI_REGISTRY}/${STABLE_OCI_REGISTRY_REPO}"
  helm_tag="${image_name}-${STABLE_BUILD_TAG}"
fi

# Change to the staging directory
cd $staging_dir || exit 1

# This codebase will probably use resfdeploy so let's just rename the manifest
# files to something that makes more sense for Helm
move_deployment() {
  mv "$1" "deployment.yaml"
  helm_repo="$(grep "calculated-image:" deployment.yaml | cut -d '"' -f 2)"
  helm_tag="$(grep "calculated-tag:" deployment.yaml | cut -d '"' -f 2)"
}
f="helm-001-ns-sa.yaml"; test -f "$f" && mv "$f" "serviceaccount.yaml"
f="helm-002-migrate.yaml"; test -f "$f" && mv "$f" "migrate.yaml"
f="helm-003-deployment.yaml"; test -f "$f" && move_deployment "$f"
f="helm-004-svc-vs-dr.yaml"; test -f "$f" && mv "$f" "service-ingress.yaml"

# Move yaml files that isn't Chart.yaml or values.yaml to the templates directory
mkdir -p templates
mv !(Chart.yaml|values.yaml|templates|.helmignore) templates

# Envsubst _helpers.tpl to fill in $NAME
CHART_NAME="$($YQ_BIN '.name' Chart.yaml)"
sed "s/{NAME}/$CHART_NAME/" templates/_helpers.tpl > templates/_helpers.tpl.new
rm -f templates/_helpers.tpl
mv templates/_helpers.tpl.new templates/_helpers.tpl

# Since the stage variable is required, make it "known" in values.yaml
chmod 777 values.yaml
echo "# The stage variable should be set to correct environment during deployment" >> values.yaml
echo "stage: prod" >> values.yaml

# The database connection variables are standardized, add here and make it known
# Only add the database variables for non-frontend charts
# todo(mustafa): add a better way to determine this
# tip: deploy.jsonnet already "knows" if a service requires a database or not
if [[ "$CHART_NAME" != *-frontend ]]; then
  echo "# For database connection" >> values.yaml
  echo "# Set postgresqlHostPort if you use initdb" >> values.yaml
  echo "postgresqlHostPort: null" >> values.yaml
  echo "# Set databaseUrl if you don't use initdb" >> values.yaml
  echo "databaseUrl: null" >> values.yaml
fi

# Service account name can also be customized
echo "# The service account name can be customized" >> values.yaml
echo "serviceAccountName: null" >> values.yaml

# Set default image values
${YQ_BIN} -i '.image.repository = '"\"$helm_repo\"" values.yaml
${YQ_BIN} -i '.image.tag = '"\"$helm_tag\"" values.yaml
${YQ_BIN} -i '.replicas = 1' values.yaml

# Helm package the chart
${HELM_BIN} package . > /dev/null 2>&1
mv ./*.tgz "$tarball_file_path"
