---
title: "Service Entry"
description: "Service Entry deployment documentation"
icon: "package"
order: 100
componentName: "kubernetesserviceentry"
---

# Kubernetes Service Entry

Provision an Istio `ServiceEntry` -- the mesh primitive that adds an external or
otherwise-unknown service into Istio's service registry, so mesh workloads can route
to it, apply traffic policy and telemetry against it, and verify its TLS identity.

## What Gets Created

- A namespaced `networking.istio.io/v1` `ServiceEntry` custom resource.
- `hosts` plus an optional combination of `addresses`, `ports`, `location`,
  `resolution`, and either static `endpoints` or a `workload_selector`.

## Prerequisites

- Istio CRDs installed on the cluster (`KubernetesIstioBaseCrds`).
- A running Istio control plane, istiod (`KubernetesIstio`), to program the registry.
- The target namespace (`KubernetesNamespace`).

## Quick Start

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesServiceEntry
metadata:
  name: external-payments-api
spec:
  namespace:
    value: payments
  hosts:
    - api.stripe.com
  location: MESH_EXTERNAL
  resolution: DNS
  ports:
    - number: 443
      name: https
      protocol: TLS
```

```bash
openmcf apply -f serviceentry.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `namespace` | reference | Namespace the ServiceEntry is created in. |
| `hosts` | list | Hosts the entry matches; at least one, no bare `*`. |

### Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `addresses` | list | Virtual IPs / CIDR prefixes (CIDR only with NONE/STATIC resolution). |
| `ports` | list | Exposed ports (`number`, `name`, `protocol`, `target_port`); name + number unique. |
| `location` | string | `MESH_EXTERNAL` (default) or `MESH_INTERNAL`. |
| `resolution` | string | `NONE` (default), `STATIC`, `DNS`, `DNS_ROUND_ROBIN`. |
| `endpoints` | list | Static backing endpoints; mutually exclusive with `workload_selector`. |
| `export_to` | list | Namespaces the service is visible to (default all). |
| `subject_alt_names` | list | SANs verified on the server certificate. |
| `workload_selector.labels` | map | In-mesh workloads (MESH_INTERNAL); mutually exclusive with `endpoints`. |

## Stack Outputs

| Output | Description |
|--------|-------------|
| `service_entry_name` | Name of the created ServiceEntry (equals metadata.name). |
| `namespace` | Namespace the ServiceEntry was created in. |

## Related Components

- [Kubernetes Istio](kubernetesistio)
- [Kubernetes Istio Base CRDs](kubernetesistiobasecrds)
- [Kubernetes Namespace](kubernetesnamespace)
