---
title: "Tune an outbound cluster (MERGE)"
description: "Merge low-level Envoy cluster settings -- connection timeout, TCP keepalive, circuit-breaker internals, and other knobs not exposed by `DestinationRule` -- onto the CDS cluster a sidecar generates..."
type: "preset"
rank: "02"
presetSlug: "02-outbound-cluster-merge"
componentSlug: "envoy-filter"
componentTitle: "Envoy Filter"
provider: "kubernetes"
icon: "package"
order: 2
---

# Tune an outbound cluster (MERGE)

Merge low-level Envoy cluster settings -- connection timeout, TCP keepalive, circuit-breaker
internals, and other knobs not exposed by `DestinationRule` -- onto the CDS cluster a sidecar
generates for a specific upstream service.

## When to Use

- You need an Envoy cluster setting that `DestinationRule` does not surface (e.g. a specific
  `upstream_connection_options`, `http2_protocol_options` detail, or transport-socket tweak).
- You are tuning the outbound path from a specific app to a specific upstream and a typed API
  does not cover the field.

## Key Configuration Choices

- **`apply_to: CLUSTER` + `context: SIDECAR_OUTBOUND`** -- patches the outbound cluster the
  sidecar builds for the upstream.
- **`match.cluster.service`** -- the upstream's fully-qualified service name (for a
  service-entry host, this is the service entry's host). Leave other cluster fields empty to
  match any port/subset; set `subset` to scope to a DestinationRule subset.
- **`patch.operation: MERGE`** -- proto-merges your `value` onto the generated cluster, leaving
  everything else intact (use `REPLACE` only when you mean to supply the cluster in its
  entirety; that is rarely what you want).
- **`workload_selector`** -- scopes the patch to the client workloads that should get the tuned
  behavior. Prefer this over a namespace-wide filter.

## Prerequisites

- The Istio CRDs are installed (`KubernetesIstioBaseCrds`).
- istiod is running and the client workloads have sidecars (`KubernetesIstio`).
- The target namespace exists (`KubernetesNamespace`).

## Placeholders to Replace

| Placeholder | Description |
|-------------|-------------|
| `<namespace>` | Namespace of the client workloads (e.g. `payments`). |
| `<app-label>` | The `app` label value selecting the client workloads. |
| `<target-service-fqdn>` | The upstream service FQDN, e.g. `reviews.default.svc.cluster.local`. |

This is an expert-only escape hatch. Express what you can through a `DestinationRule` first;
use this MERGE patch only for cluster fields the typed API does not expose.
