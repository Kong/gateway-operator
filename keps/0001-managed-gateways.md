---
title: Managed Gateways
status: provisional
---

# Managed Gateways

<!-- toc -->
- [Summary](#summary)
- [Motivation](#motivation)
  - [Goals](#goals)
  - [Non-Goals](#non-goals)
- [Proposal](#proposal)
  - [Design](#design)
    - [API Resources](#api-resources)
    - [Controllers](#controllers)
    - [Development Environment](#development-environment)
  - [Test plan](#test-plan)
  - [Graduation Criteria](#graduation-criteria)
- [Production Readiness](#production-readiness)
- [Alternatives](#alternatives)
<!-- /toc -->

## Summary

Historically the [Kong Kubernetes Ingress Controller (KIC)][kic] was used to
manage routing traffic to a back-end [Kong Gateway][kong] and both were deployed
using [Helm Chart][charts].

The purpose of this proposal is to suggest an alternative deployment mechanism
founded on the [operator pattern][operators] which would allow Kong Gateways to
be provisioned in a dynamic and Kubernetes-native way, as well as enabling
automation of Kong cluster operations and management of the Gateway lifecycle.

[kic]:https://github.com/kong/kubernetes-ingress-controller
[kong]:https://github.com/kong/kong
[charts]:https://github.com/kong/charts
[operators]:https://kubernetes.io/docs/concepts/extend-kubernetes/operator/

## Motivation

- streamline deployment and operations of Kong on Kubernetes
- configure and manage Kong on Kubernetes using CRDs and the Kubernetes API (as opposed to Helm templating)
- easily manage and view multiple deployments of the Gateway in a single cluster
- automate historically manual cluster operations for Kong (such as upgrades and scaling)

### Goals

- create a foundational [golang-based][gosdk] operator for Kong
- enable deploying Kong with Kubernetes [Gateway][gwapis] resources
- enable deploying the KIC configured to manage any number `Gateways` (resolves historical [KIC#702][kic702])
- enable automated canary upgrades of Kong (and KIC)
- stay in spec with [Operator Framework][ofrm] standards so that we can be published on [operatorhub][ohub]
- provide easy defaults for deployment while also enabling power users

[gosdk]:https://sdk.operatorframework.io/docs/building-operators/golang/quickstart/
[gwapis]:https://github.com/kubernetes-sigs/gateway-api
[kic702]:https://github.com/Kong/kubernetes-ingress-controller/issues/702
[ofrm]:https://operatorframework.io/
[ohub]:https://operatorhub.io/

## Proposal

### Design

The operator will introduce some new [Custom Resource Definitions (CRDs)][crds]
in addition to building upon resources already defined in [Gateway API][gwapi]:

- `DataPlane`
- `ControlPlane`
- `Gateway` (upstream)
- `GatewayClass` (upstream)
- `GatewayConfiguration`

The `DataPlane` resource will create and manage the lifecycle of `Deployments`,
`Secrets` and `Services` relevant to the [Kong Gateway][kong]. By creating a
`DataPlane` resource on the cluster you can deploy and configure an edge proxy
for ingress traffic. It is expected that this API will be used predominantly as
an implementation detail for `Gateway` resources, and not be used directly by
end-users (though we'll make it possible to do so).

The `ControlPlane` resource will create and manage the lifecycle of `Deployments`,
`Secrets` and `Services` relevant to the [Kong Kubernetes Ingress Controller][kic].
Creating a `ControlPlane` enables the use of `Ingress` and `HTTPRoute` APIs for
the any number of `DataPlanes` the `ControlPlane` is configured to managed.
Similar to the `DataPlane` resource above this is not designed to be an end-user
API.

The upstream `Gateway` and `GatewayClass` resources can be used to create and
manage the lifecycle of both `DataPlanes` and `ControlPlanes`. By default the
creation of a `Gateway` (which is configured with the relevant `GatewayClass`)
results in a default `ControlPlane` and `DataPlane`, with a single proxy
instance listening for ingress traffic and configurable via Kubernetes APIs
(e.g. `Ingress`, `HTTPRoute`, `TCPRoute`, `UDPRoute`, e.t.c.).

Finally, the `GatewayConfiguration` resource allows for implementation specific
configuration of `Gateways`. Some things like listener configuration can be
defined in the `Gateway` resource and that would work across all the
implementations outside of Kong, but things like configuring the max number of
nginx worker processes may not.

[crds]:https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/
[gwapi]:https://github.com/kubernetes-sigs/gateway-api
[kong]:https://github.com/kong/kong
[kic]:https://github.com/kong/kubernetes-ingress-controller

#### Development Environment

We will fundamentally rely on the [Operator SDK][osdk] as the tooling for the
project using the [Golang Operator][gopr] paradigm which will allow us to
quickly and easily ship package updates for [Operator Hub][ohub].

In previous projects such as the KIC we had used [Kubebuilder][kb] which is a
subset of Operator SDK, but the Operator SDK gives us additional features on
top of this mainly for packaging and distributing our operator via the Operator
Hub which is an explicit goal of the project.

At the time of writing the Operator SDK was under [CNCF][cncf] governance as an
[incubating][prjx] project gave us confidence that we can rely on it in the
long term.

[osdk]:https://github.com/operator-framework/operator-sdk
[gopr]:https://sdk.operatorframework.io/docs/building-operators/golang/quickstart/
[ohub]:https://operatorhub.io/
[prjx]:https://www.cncf.io/projects/

### Test Plan

Testing for this new operator will be performed similarly to what's already
performed in the [KIC][kic] which will include:

- unit tests for all Go packages using [go tests][gotest]
- integration tests using the [Kong Kubernetes Testing Framework (KTF)][ktf]
- e2e tests using [KTF][ktf]

All tests should be able to run locally using `go test` or for integration and
e2e tests using a local system Kubernetes deployment like [Kubernetes in Docker
(KIND)][kind].

[kic]:https://github.com/kong/kubernetes-ingress-controller
[gotest]:https://pkg.go.dev/testing
[ktf]:https://github.com/kong/kubernetes-testing-framework
[kind]:https://github.com/kubernetes-sigs/kind

### Graduation Criteria

The following milestones are considered prerequisites for a generally available
`v1` release of the Gateway Operator:

#### Milestone - Managed Gateway Fundamentals

This milestone covers the basic functionality and automation needed for a
simple, traditional edge proxy deployment with Kubernetes API support which is
instrumented with `Gateway` resources (and optionally `GatewayConfigurations`).

This milestone corresponds with [Operator Capabilities Level 1 "Basic
Install][ocap].

View this milestone and all its issues on Github [here][gom1].

[ocap]:https://operatorframework.io/operator-capabilities/
[gom1]:https://github.com/Kong/gateway-operator/milestone/1

#### Milestone - Manual Canary Upgrades

This milestone covers the functionality neededed to perform a canary upgrade
of a `Gateway`.

This milestone corresponds with [Operator Capabilities Level 2 "Seamless
Upgrades"][ocap].

View this milestone and all its issues on Github [here][gom8].

[ocap]:https://operatorframework.io/operator-capabilities/
[gom8]:https://github.com/Kong/gateway-operator/milestone/8

#### Milestone - Automated Upgrades

This milestone covers automating canary upgrades so that users can define
a strategy for automatic upgrades and testing automatically completes the
upgrade.

This milestone corresponds with [Operator Capabilities Level 3 "Full
Lifecycle"][ocap].

View this milestone and all its issues on Github [here][gom9].

[ocap]:https://operatorframework.io/operator-capabilities/
[gom9]:https://github.com/Kong/gateway-operator/milestone/9

#### Milestone - Backup & Restore

This milestone covers backing up and restoring the state of a `Gateway` so that
it can be restored to a new cluster, or a previous state can be restored.

This milestone corresponds with [Operator Capabilities Level 3 "Full
Lifecycle"][ocap].

View this milestone and all its issues on Github [here][gom12].

[ocap]:https://operatorframework.io/operator-capabilities/
[gom12]:https://github.com/Kong/gateway-operator/milestone/12

#### Milestone - Monitoring

Integration with [Prometheus][prometheus] to provide monitoring and insights
and reporting of failures (such as upgrade or backup failures).

This milestone corresponds with [Operator Capabilities Level 4 "Deep
Insights"][ocap].

View this milestone and all its issues on Github [here][gom14].

[prometheus]:https://prometheus.io/
[ocap]:https://operatorframework.io/operator-capabilities/
[gom14]:https://github.com/Kong/gateway-operator/milestone/14

#### Milestone - Autoscaling

This milestone covers automating `Gateway` scaling to scale the number of
underlying instances to support growing traffic dynamically.

This milestone corresponds with [Operator Capabilities Level 5 "Auto
Pilot"][ocap].

View this milestone and all its issues on Github [here][gom15].

[ocap]:https://operatorframework.io/operator-capabilities/
[gom15]:https://github.com/Kong/gateway-operator/milestone/15

#### Milestone - Documentation

This milestone covers getting full and published documentation for the Gateway
Operator published on the [Kong Documentation Site][kongdocs] under it's own
product listing.

View this milestone and all its issues on Github [here][gom16].

[kongdocs]:https://docs.konghq.com
[gom16]:https://github.com/Kong/gateway-operator/milestone/16

## Production Readiness

Production readiness of this operator is marked by the following requirements:

- [ ] All milestones of the above `Graduation Criteria` have been completed
- [ ] Unit, integration and E2E tests are present at a high level of coverage

## Alternatives

### Blue/Green Upgrades

Originally we considered automated blue/green upgrades for our `v1` release of
the operator, however we decided that canary upgrades would be sufficient for
the initial release as this would reduce scope while also focusing on what we
expected to be the more commonly used strategy. We will be able to revisit
whether to add blue/green upgrades as part of a later iteration and separate KEP.

### PostGreSQL Mode

For the first iteration we are _only_ supporting DBLESS mode deployments of
Kong as this is the preferred operational mode on Kubernetes and PostGreSQL
mode adds a lot of burden and complexity.
