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

- streamline deployment of Kong on Kubernetes
- configure and manage Kong on Kubernetes using CRDs and the Kubernetes API (as opposed to Helm templating)
- easily manage and view multiple deployments of the Gateway in a single cluster
- automate historically manual cluster operations for Kong (such as upgrades)

### Goals

- create a foundational [golang-based][gosdk] operator for Kong
- enable deploying Kong with Kubernetes [Gateway][gwapis] resources
- enable deploying the KIC configured to manage any number `Gateways` (resolves historical [KIC#702][kic702])
- enable automated blue/green upgrades of Kong (and KIC)
- stay in spec with [Operator Framework][ofrm] standards so that we can be published on [operatorhub][ohub]

[gosdk]:https://sdk.operatorframework.io/docs/building-operators/golang/quickstart/
[gwapis]:https://github.com/kubernetes-sigs/gateway-api
[kic702]:https://github.com/Kong/kubernetes-ingress-controller/issues/702
[ofrm]:https://operatorframework.io/
[ohub]:https://operatorhub.io/

### Non-Goals

- in spite of desire we are avoiding automatic rollbacks for `Gateways` within
  this scope due to complexity. Such a feature will instead require it's own
  separate KEP and design. See the alternatives section for more information.

## Proposal

### Design

The operator will introduce some new [Custom Resource Definitions (CRDs)][crds]
in addition to building upon `Gateway` API:

- `Gateway` - used as a central resource for managing the configuration and
  lifecycle of data-plane components (e.g. kong gateway `Deployments`)
- `GatewayConfiguration` - used to provide configuration options to `Gateway`
  resources and manage versioning.
- `ControlPlane` - central resource for managing the configuration and lifecycle
  of control-plane components (e.g. ingress controller `Deployments`)

The operator will employ several controllers which transact the lifecycle
management and workflows of `ControlPlane` and `Gateway` resources:

- `data-plane-deployment-controller`
- `data-plane-upgrade-controller`
- `data-plane-health-controller`
- `control-plane-deployment-controller`
- `control-plane-upgrade-controller`
- `control-plane-health-controller`

In the following sections we'll dive into the characteristics and details of
each of the components and their responsibilities.

[gosdk]:https://sdk.operatorframework.io/docs/building-operators/golang/quickstart/
[crds]:https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/

#### API Resources

##### Gateway/GatewayClass/GatewayConfiguration

The `Gateway` resource will be used to declaratively deploy a [Kong
Gateway][kong]. `Gateway` itself will include the addresses, listeners, and
protocols that define how the `Gateway` can be interacted with.

A new [GatewayClass Parameters CRD][gwparam] will be created to allow
configuration passthrough to the `Pods` of `Deployments` created on behalf of
`Gateway` resources:

```golang
type GatewayConfiguration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GatewayConfigurationSpec   `json:"spec,omitempty"`
	Status GatewayConfigurationStatus `json:"status,omitempty"`
}

type GatewayConfigurationSpec struct {
	// ContainerImage indicates the image name to use for the underlying Gateway
	// Deployment and pairs with the spec.Version option to allow specifying the
	// version of the image to use.
	//
	// +optional
	// +kubebuilder:default=DefaultGatewayContainerImage
	ContainerImage *string `json:"containerImage,omitempty"`

	// Version indicates the desired version of the ControlPlane which will also
	// correspond to the tag used for the ContainerImage.
	//
	// Not available when AutomaticUpgrades is in use.
	//
	// If omitted, the latest stable version will be used.
	//
	// +optional
	Version *string `json:"version,omitempty"`

	// Env provides environment variables that will be distributed to any Gateway
	// which is attached to this Configuration.
	//
	// +optional
	Env []corev1.EnvVar `json:"env,omitempty"`

	// Env provides environment variables that will be distributed to any Gateway
	// which is attached to this Configuration with the values for for those
	// environment variables coming from a specified source.
	//
	// +optional
	EnvFrom []corev1.EnvFromSource `json:"envFrom,omitempty"`
}
```

The above `GatewayConfiguration` effectively describes the `PodSpec` for any
`Deployments` created for the `Gateways`. `Env` provides direct access to all
the tunable configuration options that [Kong][kong] itself has available.

Notably, the `AutomaticRollback` option is not present here (seen on the
upcoming `ControlPlane` type) as this was deemed to be a significant and
complex enough problem to warrant it's own KEP and project, see the non-goals
and alternatives sections for more information.

The following manifest provides an example of how a `Gateway` can be created
which will serve HTTP traffic on port 80, have a specified admin password and
use 4 worker processes per container:

```
kind: Secret
apiVersion: v1
metadata:
  name: kong-admin-password
type: Opaque
data:
  password: bWFkZSB5b3UgbG9vawo=
---
kind: GatewayConfiguration
apiVersion: operator.konghq.com/v1alpha1
metadata:
  name: kong-gateway-config
spec:
  containerImage: kong/kong
  version: 2.7.1
  env:
  - name: KONG_NGINX_WORKER_PROCESSES
    value: 4
  - name: KONG_ADMIN_PASSWORD
    valueFrom:
      secretKeyRef:
        name: kong-admin-password
        key: password
---
kind: GatewayClass
apiVersion: gateway.networking.k8s.io/v1alpha2
metadata:
  name: kong
spec:
  controllerName: konghq.com/operator
  parametersRef:
    group: operator.konghq.com/v1alpha1
    kind: GatewayConfig
    name: kong-gateway-config
---
kind: Gateway
apiVersion: gateway.networking.k8s.io/v1alpha2
metadata:
  name: kong-gateway-1
spec:
  gatewayClassName: kong
  listeners:
  - name: http
    protocol: HTTP
    port: 80
```

If more `Gateway` resources are created and attached to the `GatewayClass` they
will inherit the same `GatewayConfiguration`.

[gwparam]:https://gateway-api.sigs.k8s.io/v1alpha2/api-types/gatewayclass/#gatewayclass-parameters
[kong]:https://github.com/kong/kong

##### ControlPlane Resource

The new `ControlPlane` resource represents the KIC `Deployment` and its
configuration options, prominently including versioning information:

```golang
type ControlPlane struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ControlPlaneSpec   `json:"spec,omitempty"`
	Status ControlPlaneStatus `json:"status,omitempty"`
}

type ControlPlaneSpec struct {
	// GatewayClass indicates the Gateway resources which this ControlPlane
	// should be responsible for configuring routes for (e.g. HTTPRoute,
	// TCPRoute, UDPRoute, TLSRoute, e.t.c.).
	//
	// Required for the ControlPlane to have any effect: at least one Gateway
	// must be present for configuration to be pushed to the data-plane and
	// only Gateway resources can be used to identify data-plane entities.
	//
	// +kubebuilder:default=DefaultGatewayClass
	GatewayClass *gatewayv1alpha2.ObjectName `json:"gatewayClass"`

	// IngressClass enables support for the older Ingress resource and indicates
	// which Ingress resources this ControlPlane should be responsible for.
	//
	// Routing configured this way will be applied to the Gateway resources
	// indicated by GatewayClass.
	//
	// If omitted, Ingress resources will not be supported by the ControlPlane.
	//
	// +optional
	// +kubebuilder:default=DefaultIngressClass
	IngressClass *string `json:"ingressClass,omitempty"`

	// ContainerImage indicates the image name to use for the underlying
	// ControlPlane Deployment and pairs with the spec.Version option to allow
	// specifying the version of the image to use.
	//
	// +optional
	// +kubebuilder:default=DefaultControlPlaneContainerImage
	ContainerImage *string `json:"containerImage,omitempty"`

	// Version indicates the desired version of the ControlPlane which will also
	// correspond to the tag used for the ContainerImage.
	//
	// Not available when AutomaticUpgrades is in use.
	//
	// If omitted, the latest stable version will be used.
	//
	// +optional
	Version *string `json:"version,omitempty"`

	// Env provides environment variables that will be distributed to any
	// Deployments created for the ControlPlane.
	//
	// +optional
	Env []v1.EnvVar `json:"env,omitempty"`

	// EnvFrom provides environment variables that will be distributed to
	// any Deployments created for the ControlPlane with the values for
	// those environment variables coming from a specified source.
	//
	// +optional
	EnvFrom []v1.EnvFromSource `json:"envFrom,omitempty"`

	// AutomaticUpgrades indicates that the version of the ControlPlane should
	// be managed automatically and provides the limits (if any) of what
	// versions should be used.
	//
	// Not available when Version is in use.
	//
	// +optional
	AutomaticUpgrades *AutomaticUpgrades `json:"automaticUpgrades,omitempty"`

	// AutomaticRollback indicates that any failure to upgrade or downgrade
	// should roll back to the previously known healthy state.
	//
	// +optional
	// +kubebuilder:default=true
	AutomaticRollback *bool `json:"automaticRollback,omitempty"`
}
```

The `ContainerImage` and `Version` options are the effective tunables for
targeting a specific version of the KIC and the image can also be overridden to
allow custom builds (e.g. debugging, testing, development builds).

The `Env` section allows directly tuning all the arguments available to the KIC
and will be processed into the `PodSpec` for the underlying KIC `Deployment`.

The `AutomaticUpgrades` option enables automatic upgrades providing (optional)
upper bounds:

```golang
type AutomaticUpgrades struct {
	// MaxMajorVersion indicates the maximum major release version that automation
	// should attempt to upgrade to.
	//
	// Required if MaxMinorVersion is set.
	//
	// +optional
	MaxMajorVersion *int `json:"maxMajorVersion"`

	// MaxMinorVersion indicates the maximum minor release version that automation
	// should attempt to upgrade to.
	//
	// Required if MaxPatchVersion is set.
	//
	// +optional
	MaxMinorVersion *int `json:"maxMinorVersionVersion,

	// MaxPatchVersion indicates the maximum patch release version that automation
	// should attempt to upgrade to.
	//
	// +optional
	MaxPatchVersion *int `json:"maxPatchVersion,
}
```

TODO: `ControlPlane` status

```golang
type ControlPlaneStatus struct {
	// Conditions describe the current conditions of the Gateway.
	//
	// +optional
	// +listType=map
	// +listMapKey=type
	// +kubebuilder:validation:MaxItems=8
	// +kubebuilder:default={{type: "Scheduled", status: "Unknown", reason:"NotReconciled", message:"Waiting for controller", lastTransitionTime: "1970-01-01T00:00:00Z"}}
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}
```

#### Controllers

##### DataPlane Deployment Controller

The `data-plane-deployment-controller` is responsible for managing the
lifecycle of `Gateway` objects by creating and managing [Kong
Gateways][kong].

- create proxy `Deployment` and all supporting objects
- apply overlay configurations to the proxy
- update `Deployment` when `Gateway` updates
- update `Deployment` when overlay `Secrets` update
- delete `Deployment` when `Gateway` deletes

[kong]:https://github.com/kong/kong

##### DataPlane Upgrade Controller

The `data-plane-upgrade-controller` is responsible for both the upgrades
and _downgrades_ of versions for the `Gateway`. Notable responsibilities
include:

- handle upgrades of the Kong `Deployment`
- handle downgrades of the Kong `Deployment`
- manage version resolution for `AutomaticUpgrades`

**TODO**: one of the biggest downsides of using `Gateway` directly is that we
          don't have much flexibility for the `status` specification like we do
          for the `ControlPlane` API. We're actively pursuing improvements in
          this regard, see the alternatives section for more details.

##### DataPlane Health Controller

The `data-plane-health-controller` is responsible for verifying the health
of the underlying Kong Gateway `Deployment` and other sub-components. Notable
responsibilities include:

- translating failures into status phase and conditions updates
- submitting success and failure events for the `Gateway` lifecycle
- submitting monitoring notifications for integrating solutions (e.g.
  prometheus)

**TODO**: See the TODO from the above section: the status and conditions
          available in `Gateway` today are fairly limited for our purposes:
          We're pursuing improvements upstream.

##### ControlPlane Deployment Controller

The `control-plane-deployment-controller` is responsible for the lifecycle of
`Deployment` objects on behalf of `ControlPlanes`. Notable responsibilities
include:

- create KIC `Deployment` and all supporting objects for new `ControlPlanes`
- apply overlay configurations to the KIC
- update `Deployment` when `ControlPlane` updates
- update `Deployment` when overlay `Secret` configs update
- delete `Deployment` when `ControlPlane` deletes

##### ControlPlane Upgrade Controller

The `control-plane-upgrade-controller` is responsible for both the upgrades
and _downgrades_ of versions for the `ControlPlane`. Notable responsibilities
include:

- handle upgrades of the KIC `Deployment`
- handle downgrades of the KIC `Deployment`
- manage version resolution for `AutomaticUpgrades`

###### Additional Logic Notes

Upgrades are triggered by the `spec.Version` being set to a version higher
than the `status.CurrentVersion` in the case of manual upgrades (or upgrades
from a 3rd party source). When triggered this way the upgrade will be
performed even if the `ControlPlane` is in a `Failed` state.

If `AutomaticUpgrades` are enabled then new versions are checked at a regular
interval and resolved according to the Major, Minor and Patch versions and
upgrades are performed automatically. The same validation that the webhook
uses for upgrades will be performed for these "transient versions" and if the
upgrade can not be performed (but fits within the limits of the provided
`AutomaticUpgrades` configuration) the `ControlPlane` is moved into a
`FailedUpgrade` phase and conditions, events and notifications are emitted.

The upgrade controller is responsible for marking the `status.Phase` of the
`ControlPlane` as `Upgrading` when an upgrade is performed, and
`UpgradeComplete` when the workflow is finished, but notably the health
controller is responsible for eventually moving it to `Ready` or `Failed`.

Downgrades are triggered by the `spec.Version` being set to a version lower
than the `status.CurrentVersion` in the case of manual downgrades. When
triggered this way the downgrade will be performed even if the `ControlPlane`
is in a `Failed` state.

The upgrade controller is responsible for marking the `status.Phase` of the
`ControlPlane` as `Upgrading` when an upgrade is performed, and
`UpgradeComplete` when the workflow is finished, but notably the health
controller is responsible for eventually moving it to `Ready` or `Failed`.
Similar is true for the corresponding `Downgrading` and `DowngradeComplete`.

If `AutomaticRollbacks` are enabled, either an upgrade or a downgrade will be
performed whenever a `FailedUpgrade` or `FailedDowngrade` state is reached,
and the target version will be the `status.lastHealthyVersion`.

When the `status.CurrentVersion` is the same as the `status.LastHealthyVersion`
after an upgrade or downgrade failed, this indicates that the failure occured
prior to the workflow and the version never actually changed so a rollback will
not be required, but status messaging events and notifications will be emitted
to indicate the problem to a human operator.

##### ControlPlane Health Controller

The `control-plane-health-controller` is responsible for verifying the health
of the underlying KIC `Deployment` and other subcomponents. Notable
responsibilities include:

- translating failures into status phase and conditions updates
- submitting success and failure events for the `ControlPlane` lifecycle
- submitting monitoring notifications for integrating solutions (e.g.
  prometheus)

If the `ControlPlane` is in the `DeployComplete` phase (meaning initial
deployment has finished) then the health controller marks it `Ready` upon
health check success or `Failed` if it never becomes healthy. Similarly it
resolves `UpgradeComplete` and `DowngradeComplete` to `Ready`, or if those
fail `UpgradeFailed` and `DowngradeFailed` respectively.

In the very unfortunate case that an upgrade fails, a rollback occurs and then
the _rollback fails as well_ the health controller is responsible for
submitting clear and present status condition messages describing the series
of failures, creating `Events`, and submitting monitoring notifications.

#### Validating Webhook

In addition to controllers we'll have a validating webhook for verifying the
contents and correctness of objects as they are added or updated in the API:

- upgrades and downgrades are validated with `Secret` + `spec.Version`
  to enable us to deny upgrades that aren't possible automatically (this
  will generally only be major version upgrades)
- any unsupported options or configurations in the interim before GA

Most `Gateway` validation will be provided by the [upstream webhook][gwhook]
which is where we should be contributing new features.

[gwhook]:https://github.com/kubernetes-sigs/gateway-api/tree/master/cmd/admission

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

#### Manual Upgrades/Downgrades

While the `ControlPlane` and `Gateway` are intended to have functionality for
automatic upgrades, it will not always be the case that an automatic upgrade is
possible, the following situations will require manual intervention:

- `ControlPlane`/`Gateway` upgrades with breaking changes
- 3rd party integrations have breaking changes
- CRD has breaking changes

To help handle validation a matrix of version compatibility will be codified
into the operator such that controller logic can make simple assessments, e.g.:

```golang
newControlPlaneVersion := controlPlane.Version
newGatewayVersion := gateway.Version
if isCompatible(newControlPlaneVersion, newGatewayVersion) {
```

This validation will always be run in the controllers and result in status
updates for incompatibilities (as defined above in the
`control-plane-upgrade-controller` and `data-plane-upgrade-controller`) and
subsets of it may also be run in the validating webhook to provide more clear
and upfront feedback about an incompatibility.

For this scope `Gateways` will not support automatic rollbacks _at all_ so the
mechanism is similar: any automatic upgrade configuration needs to be removed
and then the upgrade/downgrade can be triggered manually by changing the
version in the `GatewayConfiguration`.

##### Upgrading for ControlPlane or Gateway breaking changes

Making an upgrade for breaking changes of either the `ControlPlane` or
`Gateway` requires the following steps:

1. if defined, any automatic upgrade configurations are removed
2. according to the release notes for the intended version, configuration
   options are adjusted to match what is supported.
3. the version of the `ControlPlane` or `Gateway` is explicitly set to the
   new version.
4. the upgrade is processed by the relevant upgrade controller
4. once the status phase has reached `UpgradeComplete` it can be reconfigured
   for `AutomaticUpgrades`.

##### 3rd party breaking changes

If a 3rd party resource for integration needs to be upgraded to support newer
versions of that software, this is where the _operator itself_ will receive
updates to be able to produce the new resources and configuration.

##### CRD breaking changes

Breaking changes in CRDs should be ultimately rare as the vast majority of new
versions should maintain migration code from the previous ones (tooling for this
is provided for in tools like `operator sdk`). The operator does not manage
CRDs, so in the rare case that this occurs the human operator will be
responsible for taking backups of all `ControlPlanes` and `Gateways` and other
related resources, tearing down the environment completely during a scheduled
maintenance window, surgically converting all resources and then re-applying
them to the cluster. It sounds awful, but again this should _basically never
happen_ and the reality is that if you have to delete a CRD you'll have to
delete the resources for it.

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

The maturity of this project will be marked by several milestones which define
specific goals on the road to GA.

#### Milestone 1 - Gateway Deployments

This milestone corresponds with [Operator Capabilities Level 1 "Basic
Install][ocap].

- [ ] the operator can cleanly create and provision `Gateway` resources and
      apply provided configuration options (e.g. EnvVars)
- [ ] the operator can cleanly tear down `Gateways`

[ocap]:https://operatorframework.io/operator-capabilities/

#### Milestone 2 - ControlPlane Deployments

This milestone corresponds with [Operator Capabilities Level 1 "Basic
Install][ocap].

- [ ] the operator can cleanly create and provision `ControlPlane` resources
      and apply provided configuration options (e.g. EnvVars)
- [ ] the operator can cleanly tear down `ControlPlanes`

[ocap]:https://operatorframework.io/operator-capabilities/

#### Milestone 3 - Manual Upgrades

This milestone corresponds with [Operator Capabilities Level 2 "Seamless
Upgrades"][ocap].

- [ ] the operator can handle manual upgrades/downgrades of `Gateway` resources
- [ ] the operator can handle manual upgrades/downgrades of `ControlPlanes`

[ocap]:https://operatorframework.io/operator-capabilities/

#### Milestone 4 - Automated Upgrades, Rollbacks

This milestone corresponds with [Operator Capabilities Level 3 "Full
Lifecycle"][ocap]. We're notably skipping the "backups" component to
level 3 capabilities as this is something that can be provided by
other tools, such as Velero.

- [ ] automated upgrades can be configured for `Gateways`
- [ ] automated upgrades can be configured for `ControlPlanes`
- [ ] automated rollback on failure can be configured for `ControlPlanes`

#### Milestone 5 - Monitoring Integrations and Analytics

This milestone corresponds with [Operator Capabilities Level 4 "Deep
Insights"][ocap]. Initially only [Prometheus][prometheus] will be supported
as a monitoring integration and other integrations can be requested later as
they are needed.

- [ ] health metrics are provided for all controllers
- [ ] health metrics are provided for `Gateways`
- [ ] prometheus alertmanager integration

[ocap]:https://operatorframework.io/operator-capabilities/
[prometheus]:https://prometheus.io/

## Production Readiness

Production readiness of this operator is marked by the following requirements:

- All milestones of the above `Graduation Criteria` have been completed
- Unit, integration and E2E tests are present at a high level of coverage
- Documentation is present in the [Kong Documentation][kongdocs]

[kongdocs]:https://github.com/kong/docs.konghq.com

## Alternatives

### DataPlane CRD

**TODO**: this alternative might still be viable if we can't get significant
          upstream improvements for `Gateway` status.

While working on this KEP it was considered to add a `DataPlane` API which
would be similar in scope to `ControlPlane` and allow for features like
scaling up and down sets of `Gateway` objects. This may still be something
that we want to do in time, but at the time of writing it didn't seem that
doing so strongly tied in with any of our existing goals and so its being
left as a consideration for later iterations.

### Automatic Gateway Rollbacks

We had considered adding automatic `Gateway` rollbacks similarly to the
automatic `ControlPlane` rollbacks we described above, however the problem
therein is that the kong `Gateway` implementation is not always stateless (e.g.
postgres backed kong) and so this means significant complexity in implementing
such a feature. There was interest from maintainers at the time of writing and
we expected that the problem is solvable, but it's big and complex enough that
we felt it should be it's own separate KEP and project.
