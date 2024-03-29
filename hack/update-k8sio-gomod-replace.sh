#!/bin/bash
set -euo pipefail

# Script adapted from https://github.com/kubernetes/kubernetes/issues/79384#issuecomment-521493597.

VERSION=$(cat go.mod | grep "k8s.io/kubernetes v" | sed "s/^.*v\([0-9.]*\).*/\1/")
echo "Updating k8s.io go.mod replace directives for k8s.io/kubernetes@v$VERSION"

MODS=($(
    curl -sS https://raw.githubusercontent.com/kubernetes/kubernetes/v${VERSION}/go.mod |
    sed -n 's|.*k8s.io/\(.*\) => ./staging/src/k8s.io/.*|k8s.io/\1|p'
))
for MOD in "${MODS[@]}"; do
    V=$(
        go mod download -json "${MOD}@kubernetes-${VERSION}" |
        sed -n 's|.*"Version": "\(.*\)".*|\1|p'
    )
    echo "Updating go.mod replace directive for ${MOD}"
    go mod edit "-replace=${MOD}=${MOD}@${V}"
done
go get "k8s.io/kubernetes@v${VERSION}"
