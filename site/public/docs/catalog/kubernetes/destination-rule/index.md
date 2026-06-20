---
title: "Destination Rule"
description: "Destination Rule deployment documentation"
icon: "package"
order: 100
componentName: "kubernetesdestinationrule"
---

# Kubernetes Destination Rule

Provision an Istio `DestinationRule` -- the mesh primitive that tunes *how* traffic is
sent to a service after routing: load balancing, connection pools, circuit breaking
(outlier detection), client TLS origination, and named subsets.

## What Gets Created

- A namespaced `networking.istio.io/v1` `DestinationRule` custom resource.
- `host` plus an optional `traffic_policy`, `subsets`, `export_to`, and
  `workload_selector`.

## Prerequisites

- Istio CRDs installed on the cluster (`KubernetesIstioBaseCrds`).
- A running Istio control plane, istiod (`KubernetesIstio`), to apply the policy.
- The target namespace (`KubernetesNamespace`).

## Quick Start

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesDestinationRule
metadata:
  name: reviews-lb
spec:
  namespace:
    value: bookinfo
  host: reviews.bookinfo.svc.cluster.local
  traffic_policy:
    load_balancer:
      simple: LEAST_REQUEST
```

```bash
openmcf apply -f destinationrule.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `namespace` | reference | Namespace the DestinationRule is created in. |
| `host` | string | Registry host the rule applies to (prefer FQDN). |

### Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `traffic_policy` | object | Load balancing, connection pool, outlier detection, TLS, per-port overrides, tunnel, PROXY protocol. |
| `subsets` | list | Named subsets (`name`, `labels`, per-subset `traffic_policy`). |
| `export_to` | list | Namespaces the rule is visible to (default all). |
| `workload_selector.match_labels` | map | Pods/VMs the rule applies to; matched by istiod, not a foreign key. |

## Stack Outputs

| Output | Description |
|--------|-------------|
| `destination_rule_name` | Name of the created DestinationRule (equals metadata.name). |
| `namespace` | Namespace the DestinationRule was created in. |

## Related Components

- [Kubernetes Service Entry](kubernetesserviceentry)
- [Kubernetes Istio](kubernetesistio)
- [Kubernetes Istio Base CRDs](kubernetesistiobasecrds)
- [Kubernetes Namespace](kubernetesnamespace)
