# Changelog

## Table of Contents

- [v0.1.0](#v010)

## v0.1.0

**Maturity: ALPHA**

This is the initial release which includes basic functionality at an alpha
level of maturity and includes some of the fundamental APIs needed to create
gateways for ingress traffic.

### Initial Features

- `GatewayClass` support was added to delineate which `Gateway` resources the
  operator supports.
  [#22](https://github.com/Kong/gateway-operator/issues/22)
- `Gateway` support was added: used to create edge proxies for ingress traffic.
  [#6](https://github.com/Kong/gateway-operator/issues/22)
- The `ControlPlane` API was added to deploy Kong Ingress Controllers which
  can be attached to `DataPlane` resources.
  [#5](https://github.com/Kong/gateway-operator/issues/22)
- The `DataPlane` API was added to deploy Kong Gateways.
  [#4](https://github.com/Kong/gateway-operator/issues/22)
