# ------------------------------------------------------------------------------
# Builder
# ------------------------------------------------------------------------------

FROM --platform=$BUILDPLATFORM golang:1.24.0@sha256:3f7444391c51a11a039bf0359ee81cc64e663c17d787ad0e637a4de1a3f62a71 AS builder

WORKDIR /workspace
ARG GOPATH
ARG GOCACHE
# Use cache mounts to cache Go dependencies and bind mounts to avoid unnecessary
# layers when using COPY instructions for go.mod and go.sum.
# https://docs.docker.com/build/guide/mounts/
RUN --mount=type=cache,target=$GOPATH/pkg/mod \
    --mount=type=cache,target=$GOCACHE \
    --mount=type=bind,source=go.sum,target=go.sum \
    --mount=type=bind,source=go.mod,target=go.mod \
    go mod download -x

COPY cmd/main.go cmd/main.go
COPY modules/ modules/
COPY controller/ controller/
COPY pkg/ pkg/
COPY internal/ internal/
COPY Makefile Makefile
COPY .git/ .git/

ARG TARGETPLATFORM
ARG TARGETOS
ARG TARGETARCH
ARG TAG
ARG COMMIT
ARG REPO_INFO

RUN printf "Building for TARGETPLATFORM=${TARGETPLATFORM}" \
    && printf ", TARGETARCH=${TARGETARCH}" \
    && printf ", TARGETOS=${TARGETOS}" \
    && printf ", TARGETVARIANT=${TARGETVARIANT} \n" \
    && printf "With 'uname -s': $(uname -s) and 'uname -m': $(uname -m)"

# Use cache mounts to cache Go dependencies and bind mounts to avoid unnecessary
# layers when using COPY instructions for go.mod and go.sum.
# https://docs.docker.com/build/guide/mounts/
RUN --mount=type=cache,target=$GOPATH/pkg/mod \
    --mount=type=cache,target=$GOCACHE \
    --mount=type=bind,source=go.sum,target=go.sum \
    --mount=type=bind,source=go.mod,target=go.mod \
    CGO_ENABLED=0 GOOS=linux GOARCH="${TARGETARCH}" \
    TAG="${TAG}" COMMIT="${COMMIT}" REPO_INFO="${REPO_INFO}" \
    make build.operator

# ------------------------------------------------------------------------------
# Distroless (default)
# ------------------------------------------------------------------------------

# Use distroless as minimal base image to package the operator binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot@sha256:6ec5aa99dc335666e79dc64e4a6c8b89c33a543a1967f20d360922a80dd21f02 AS distroless

ARG TAG
ARG NAME="Kong Gateway Operator"
ARG DESCRIPTION="Kong Gateway Operator drives deployment via the Gateway resource. You can deploy a Gateway resource to the cluster which will result in the underlying control-plane (the Kong Kubernetes Ingress Controller) and the data-plane (the Kong Gateway)."

LABEL name="${NAME}" \
    description="${DESCRIPTION}" \
    org.opencontainers.image.description="${DESCRIPTION}" \
    vendor="Kong" \
    version="${TAG}" \
    release="1" \
    url="https://github.com/Kong/gateway-operator" \
    summary="A Kubernetes Operator for the Kong Gateway."

WORKDIR /
COPY --from=builder /workspace/bin/manager .
USER 65532:65532

ENTRYPOINT ["/manager"]
