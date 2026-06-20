---
title: "TCP Weighted Backends"
description: "A TCPRoute rule that splits raw TCP connections across two backends by weight -- the building block for a canary or blue/green rollout of a non-HTTP service. Connection rejections (for invalid..."
type: "preset"
rank: "02"
presetSlug: "02-tcp-weighted-backends"
componentSlug: "tcp-route"
componentTitle: "TCP Route"
provider: "kubernetes"
icon: "package"
order: 2
---

# TCP Weighted Backends

A TCPRoute rule that splits raw TCP connections across two backends by weight --
the building block for a canary or blue/green rollout of a non-HTTP service.
Connection rejections (for invalid backends) respect weight too.

## When to Use

- You are rolling out a new version of a TCP backend and want to shift a fraction
  of connections to it.
- You need weighted connection distribution across backends on the same listener.

## Key Configuration Choices

- **`backendRefs[].weight`** -- relative share of connections per backend,
  computed as `weight / sum(weights)`. Here 90/10 sends ~10% to the canary.
- A TCP route has no matching; all connections on the listener are split across
  the rule's weighted backends.

## Prerequisites

- The Gateway API **experimental-channel** CRDs are installed
  (`KubernetesGatewayApiCrds` with `install_channel: experimental`).
- The `Gateway` referenced in `parentRefs` exists (`KubernetesGateway`) with a
  `TCP` listener.
- The target namespace exists (`KubernetesNamespace`).
- Both backend Services exist in the route's namespace.

## Placeholders to Replace

| Placeholder | Description |
|-------------|-------------|
| `<gateway-name>` | Name of the `KubernetesGateway` this route attaches to. |
| `<stable-service>` | Name of the current (stable) backend Service. |
| `<canary-service>` | Name of the new (canary) backend Service. |
| `<service-port>` | Backend Service port. |

Tune the `weight` values to control the split; set `spec.namespace.value` to your
namespace, or replace it with a `valueFrom` reference to a `KubernetesNamespace`.
