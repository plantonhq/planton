---
title: "Envoy Filter"
description: "Envoy Filter deployment documentation"
icon: "package"
order: 100
componentName: "kubernetesenvoyfilter"
---

# Kubernetes Envoy Filter

Provision an Istio `EnvoyFilter` -- the expert-only escape hatch that patches the Envoy proxy
configuration istiod generates for selected workloads, when no first-class Istio API expresses
what you need.

> **Expert-only.** The patch body is free-form xDS JSON that istiod merges with no schema
> validation; a malformed patch can break a workload's traffic. Prefer a typed Istio API first.

## What Gets Created

- A namespaced `networking.istio.io/v1alpha3` `EnvoyFilter` custom resource.
- An attachment scope (`workload_selector` or `target_refs`, or neither) plus an ordered list
  of `config_patches` and an optional `priority`.

## Prerequisites

- Istio CRDs installed on the cluster (`KubernetesIstioBaseCrds`).
- A running Istio control plane, istiod (`KubernetesIstio`), to translate the patches.
- The target namespace (`KubernetesNamespace`).

## Quick Start

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesEnvoyFilter
metadata:
  name: outbound-timeout
spec:
  namespace:
    value: payments
  workload_selector:
    labels:
      app: checkout
  config_patches:
    - apply_to: CLUSTER
      match:
        context: SIDECAR_OUTBOUND
        cluster:
          service: reviews.default.svc.cluster.local
      patch:
        operation: MERGE
        value:
          connect_timeout: 5s
```

```bash
openmcf apply -f envoyfilter.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `namespace` | reference | Namespace the EnvoyFilter is created in. |

### Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `workload_selector.labels` | map | Pod/VM labels the patches apply to. Mutually exclusive with `target_refs`. |
| `target_refs` | list | Attach to specific resources (`group`/`kind`/`name`); max 16. Mutually exclusive with `workload_selector`. |
| `config_patches` | list | Ordered patches (`apply_to`, `match`, `patch`). An empty list is a valid no-op. |
| `priority` | int | Patch-set ordering within a context (default 0). |

## Stack Outputs

| Output | Description |
|--------|-------------|
| `envoy_filter_name` | Name of the created EnvoyFilter (equals metadata.name). |
| `namespace` | Namespace the EnvoyFilter was created in. |

## Related Components

- [Kubernetes Istio](kubernetesistio)
- [Kubernetes Istio Base CRDs](kubernetesistiobasecrds)
- [Kubernetes Namespace](kubernetesnamespace)
