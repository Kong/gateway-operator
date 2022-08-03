# Changelog

## Table of Contents

- [v0.1.0](#v010)

## v0.1.0 (unreleased)

**Maturity: ALPHA**

This is the initial release which includes basic functionality at an alpha
level of maturity and includes some of the fundamental APIs needed to create
gateways for ingress traffic.

### Initial Features

- The `GatewayConfiguration` API was added to enable configuring `Gateway`
  resources with the options needed to influence the configuration of
  the underlying `ControlPlane` and `DataPlane` resources.
  [#43](https://github.com/Kong/gateway-operator/pull/43)
- `GatewayClass` support was added to delineate which `Gateway` resources the
  operator supports.
  [#22](https://github.com/Kong/gateway-operator/issues/22)
- `Gateway` support was added: used to create edge proxies for ingress traffic.
  [#6](https://github.com/Kong/gateway-operator/issues/6)
- The `ControlPlane` API was added to deploy Kong Ingress Controllers which
  can be attached to `DataPlane` resources.
  [#5](https://github.com/Kong/gateway-operator/issues/5)
- The `DataPlane` API was added to deploy Kong Gateways.
  [#4](https://github.com/Kong/gateway-operator/issues/4)
- The operator manages certificates for control and data plane communication
  and configures mutual TLS between them. It cannot yet replace expired
  certificates.
  [#103](https://github.com/Kong/gateway-operator/issues/103)

[v0.1.0]: https://github.com/Kong/gateway-operator/tree/v0.1.0
