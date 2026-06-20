---
title: "Circuit Breaking & Outlier Detection"
description: "The canonical DestinationRule: protect a service by capping connection-pool size and ejecting hosts that keep returning errors. This is how you stop a struggling backend from taking down its callers..."
type: "preset"
rank: "01"
presetSlug: "01-circuit-breaking-outlier-detection"
componentSlug: "destination-rule"
componentTitle: "Destination Rule"
provider: "kubernetes"
icon: "package"
order: 1
---

# Circuit Breaking & Outlier Detection

The canonical DestinationRule: protect a service by capping connection-pool size and
ejecting hosts that keep returning errors. This is how you stop a struggling backend from
taking down its callers -- the mesh sheds load and routes around unhealthy endpoints
automatically.

## When to Use

- A backend service occasionally returns 5xx errors or hangs, and you want unhealthy pods
  removed from the load-balancing pool until they recover.
- You want to bound the concurrency and queue depth a client sidecar will push at an
  upstream service (classic circuit breaking).

## Key Configuration Choices

- **`traffic_policy.load_balancer.simple: LEAST_REQUEST`** -- favors the least-busy host;
  the recommended default over `ROUND_ROBIN`.
- **`connection_pool.tcp.max_connections`** / **`http.http2_max_requests`** /
  **`http.max_requests_per_connection`** -- the circuit-breaker limits. Requests beyond
  them are queued or fail fast rather than overwhelming the backend.
- **`outlier_detection.consecutive_5xx_errors`** -- how many 5xx responses eject a host.
- **`outlier_detection.interval`** / **`base_ejection_time`** -- how often hosts are scanned
  and how long an ejected host stays out (grows with repeated ejections).
- **`outlier_detection.max_ejection_percent`** -- caps how much of the pool can be ejected at
  once, so detection can't empty the pool.

## Prerequisites

- The Istio CRDs are installed (`KubernetesIstioBaseCrds`).
- istiod is running and the calling workloads have sidecars or are in the ambient mesh
  (`KubernetesIstio`).
- The target namespace exists (`KubernetesNamespace`).

## Placeholders to Replace

| Placeholder | Description |
|-------------|-------------|
| `<namespace>` | Namespace the rule is created in (e.g. `bookinfo`). |
| `<service-host>` | The service registry host to protect (e.g. `reviews.bookinfo.svc.cluster.local`). |

Pair this with a `ServiceEntry` if the host is external, and a `VirtualService` if you need
to route across subsets.
