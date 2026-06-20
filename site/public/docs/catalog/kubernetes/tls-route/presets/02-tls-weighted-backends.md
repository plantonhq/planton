---
title: "TLS Weighted Backends"
description: "A single TLSRoute rule that splits passthrough TLS connections across two backends by weight -- the building block for a canary or blue/green rollout of a TLS-terminating service. Because a TLSRoute..."
type: "preset"
rank: "02"
presetSlug: "02-tls-weighted-backends"
componentSlug: "tls-route"
componentTitle: "TLS Route"
provider: "kubernetes"
icon: "package"
order: 2
---

# TLS Weighted Backends

A single TLSRoute rule that splits passthrough TLS connections across two
backends by weight -- the building block for a canary or blue/green rollout of a
TLS-terminating service. Because a TLSRoute permits exactly one rule, traffic
splitting is expressed entirely through multiple weighted `backendRefs` within
that rule.

## When to Use

- You are rolling out a new version of a TLS-terminating backend and want to
  shift a fraction of connections to it.
- You need weighted distribution across backends for the same SNI hostname.

## Key Configuration Choices

- **`backendRefs[].weight`** -- relative share of connections per backend,
  computed as `weight / sum(weights)`. Here 90/10 sends ~10% to the canary.
- **`hostnames`** -- the SNI hostname(s) this route serves.
- A TLSRoute has exactly one rule; put all weighted backends in that rule.

## Prerequisites

- The Gateway API CRDs are installed (`KubernetesGatewayApiCrds`).
- The `Gateway` referenced in `parentRefs` exists (`KubernetesGateway`) with a
  `TLS` `Passthrough` listener.
- The target namespace exists (`KubernetesNamespace`).
- Both backend Services exist in the route's namespace.

## Placeholders to Replace

| Placeholder | Description |
|-------------|-------------|
| `<gateway-name>` | Name of the `KubernetesGateway` this route attaches to. |
| `<sni-hostname>` | SNI hostname this route serves, e.g. `secure.example.com`. |
| `<stable-service>` | Name of the current (stable) backend Service. |
| `<canary-service>` | Name of the new (canary) backend Service. |

Tune the `weight` values to control the split; set `spec.namespace.value` to your
namespace, or replace it with a `valueFrom` reference to a `KubernetesNamespace`.
