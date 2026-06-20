---
title: "Envoy Gateway GatewayClass"
description: "This preset creates a GatewayClass for the Envoy Gateway controller. Envoy Gateway is a standalone Gateway API implementation built on Envoy proxy, popular for teams that want Envoy's data plane..."
type: "preset"
rank: "02"
presetSlug: "02-envoy-gateway-class"
componentSlug: "gateway-class"
componentTitle: "Gateway Class"
provider: "kubernetes"
icon: "package"
order: 2
---

# Envoy Gateway GatewayClass

This preset creates a GatewayClass for the Envoy Gateway controller. Envoy Gateway is a standalone Gateway API implementation built on Envoy proxy, popular for teams that want Envoy's data plane without a full service mesh.

## When to Use

- Your cluster runs Envoy Gateway
- You want a lightweight, Envoy-based Gateway API data plane
- You do not need a full service mesh (compare with the Istio preset)

## Key Configuration Choices

- **controllerName** (`gateway.envoyproxy.io/gatewayclass-controller`) -- the identity Envoy Gateway's controller watches for; copied verbatim from Envoy Gateway's documentation
- **No parametersRef** -- omitted for simplicity; Envoy Gateway can be tuned by referencing an `EnvoyProxy` resource via `parametersRef` (group `gateway.envoyproxy.io`, kind `EnvoyProxy`) when advanced configuration is needed
- **Cluster-scoped** -- GatewayClass is cluster-wide; no namespace is set

## Prerequisites

- Gateway API CRDs installed (`KubernetesGatewayApiCrds`)
- Envoy Gateway installed in the cluster

## Placeholders to Replace

No placeholders -- this preset is directly deployable. Rename `metadata.name` to match your naming convention (the name becomes the GatewayClass that Gateways reference via `gatewayClassName`).

To attach a custom Envoy proxy configuration, add:

```yaml
  parametersRef:
    group: gateway.envoyproxy.io
    kind: EnvoyProxy
    name: <your-envoyproxy-resource>
    namespace: <its-namespace>
```
