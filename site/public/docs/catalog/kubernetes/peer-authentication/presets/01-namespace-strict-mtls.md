---
title: "Require Strict mTLS Across a Namespace"
description: "The canonical PeerAuthentication: require mutual TLS for every workload in a namespace. With no selector, the policy is the namespace default, so all in-mesh traffic to those workloads must arrive..."
type: "preset"
rank: "01"
presetSlug: "01-namespace-strict-mtls"
componentSlug: "peer-authentication"
componentTitle: "Peer Authentication"
provider: "kubernetes"
icon: "package"
order: 1
---

# Require Strict mTLS Across a Namespace

The canonical PeerAuthentication: require mutual TLS for every workload in a
namespace. With no selector, the policy is the namespace default, so all
in-mesh traffic to those workloads must arrive over an authenticated mTLS tunnel.

## When to Use

- You want a hard security baseline for a namespace: no plaintext traffic to any
  workload reaches the sidecar.
- All callers are already on the mesh (or you have verified they are), so STRICT
  will not break legitimate plaintext clients.

## Key Configuration Choices

- **`spec.namespace`** -- the namespace the policy governs. Naming the policy
  `default` is the Istio convention for the namespace-wide policy.
- **No `selector`** -- omitting it scopes the policy to the entire namespace.
- **`mtls.mode: STRICT`** -- rejects non-mTLS connections. Use `PERMISSIVE`
  instead while migrating callers onto the mesh, then tighten to `STRICT`.

## Prerequisites

- The Istio CRDs are installed (`KubernetesIstioBaseCrds`).
- istiod is running and the namespace's workloads have sidecars or are in the
  ambient mesh (`KubernetesIstio`).
- The target namespace exists (`KubernetesNamespace`).

## Placeholders to Replace

| Placeholder | Description |
|-------------|-------------|
| `<namespace>` | Namespace whose workloads must require mTLS (e.g. `finance`). |

Set `spec.namespace.value` to your namespace, or replace it with a `valueFrom`
reference to a `KubernetesNamespace`.
