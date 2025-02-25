#!/bin/bash

# This script extracts the commit SHA from go.mod for the kong/kubernetes-configuration
# dependency and uses it to generate a kustomization.yaml file. The SHA is appended 
# as a query parameter to the CRD resource URL to ensure versioning consistency.

set -o errexit
set -o nounset
set -o pipefail

REPO_ROOT=$(dirname ${BASH_SOURCE})/..

KCONF_PACKAGE="github.com/kong/kubernetes-configuration"
RAW_VERSION=$(go list -m -f '{{ .Version }}' ${KCONF_PACKAGE})
if [[ $(echo "${RAW_VERSION}" | tr -cd '-' | wc -c) -ge 2 ]]; then
    # If there are 2 or more hyphens, extract the part after the last hyphen as 
    # that's a git commit hash (e.g. `v1.1.1-0.20250217181409-44e5ddce290d`).
    SHA_SHORT="$(echo "${RAW_VERSION}" | rev | cut -d'-' -f1 | rev)"
    KCONF_VERSION="ref=$(curl -s https://api.github.com/repos/Kong/kubernetes-configuration/commits/${SHA_SHORT} | jq -r .sha)"
else
    KCONF_VERSION="ref=${RAW_VERSION}"
fi

generate_kustomization_file() {
  local file_path=$1
  local resource_url=$2

  echo "INFO: Generating ${file_path} with kong/kubernetes-configuration version ${KCONF_VERSION}"
  echo "# Generated by scripts/generate-crd-kustomize.sh. DO NOT EDIT.
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - ${resource_url}?${KCONF_VERSION} # Version is auto-updated by the generating script." > "${file_path}"
}

generate_kustomization_file \
  "${REPO_ROOT}/config/crd/kustomization.yaml" \
  "https://github.com/kong/kubernetes-configuration/config/crd/gateway-operator"
