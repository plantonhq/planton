---
title: "Weighted Canary Split"
description: "Send most traffic to a stable backend and a small slice to a canary, using backend weights. The Gateway distributes requests in proportion to each backend's `weight` (here 90/10), which is the..."
type: "preset"
rank: "02"
presetSlug: "02-weighted-canary"
componentSlug: "http-route"
componentTitle: "HTTP Route"
provider: "kubernetes"
icon: "package"
order: 2
---

# Weighted Canary Split

Send most traffic to a stable backend and a small slice to a canary, using
backend weights. The Gateway distributes requests in proportion to each
backend's `weight` (here 90/10), which is the foundation of progressive
delivery.

## When to Use

- You are rolling out a new version and want to shift a fraction of traffic to it.
- You want a simple, controller-native traffic split without a service mesh.
- You plan to adjust the weights over time (for example 90/10 -> 50/50 -> 0/100).

## Key Configuration Choices

- **Two `backendRefs` with weights** -- traffic is split as `weight / sum(weights)`, so 90 and 10 yield a 90%/10% split.
- **Same hostname and path** -- both backends serve identical match conditions; only the weight differs.
- **Adjusting weights** -- promote the canary by raising its weight and lowering the stable weight; weights need not sum to 100.

## Prerequisites

- The Gateway API CRDs are installed (`KubernetesGatewayApiCrds`).
- The `Gateway` referenced in `parentRefs` exists (`KubernetesGateway`).
- The target namespace exists (`KubernetesNamespace`).
- Both backend Services exist in the route's namespace.

## Placeholders to Replace

| Placeholder | Description |
|-------------|-------------|
| `<gateway-name>` | Name of the `KubernetesGateway` this route attaches to. |
| `<app-hostname>` | Public hostname this route serves, e.g. `app.example.com`. |
| `<stable-service-name>` | Service receiving the majority of traffic. |
| `<canary-service-name>` | Service receiving the canary slice of traffic. |

Set `spec.namespace.value` to your namespace, or replace it with a `valueFrom`
reference to a `KubernetesNamespace`.
