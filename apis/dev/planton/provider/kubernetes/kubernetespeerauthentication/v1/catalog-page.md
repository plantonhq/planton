# Kubernetes Peer Authentication

Provision an Istio `PeerAuthentication` -- the mesh primitive that sets mutual TLS
(mTLS) requirements for incoming connections to your workloads. Enforce
encrypted, authenticated service-to-service traffic across a namespace or the
whole mesh, ease workloads onto the mesh with permissive mode, and override
specific ports as needed.

## What Gets Created

- A namespaced `security.istio.io/v1` `PeerAuthentication` custom resource.
- An optional workload `selector`, a workload-level `mtls.mode`, and optional
  per-port `port_level_mtls` overrides.

## Prerequisites

- Istio CRDs installed on the cluster (`KubernetesIstioBaseCrds`).
- A running Istio control plane, istiod (`KubernetesIstio`), to enforce the policy.
- The target namespace (`KubernetesNamespace`).

## Quick Start

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesPeerAuthentication
metadata:
  name: default
spec:
  namespace:
    value: finance
  mtls:
    mode: STRICT
```

```bash
planton apply -f peerauthentication.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `namespace` | reference | Namespace the policy is created in. |

### Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `selector.match_labels` | map | Pod labels selecting target workloads; omit for namespace-wide scope. |
| `mtls.mode` | string | `UNSET`, `DISABLE`, `PERMISSIVE`, or `STRICT`; omit to inherit. |
| `port_level_mtls` | map | Per-workload-port mode overrides; requires a selector. |

## Stack Outputs

| Output | Description |
|--------|-------------|
| `peer_authentication_name` | Name of the created PeerAuthentication (equals metadata.name). |
| `namespace` | Namespace the PeerAuthentication was created in. |

## Related Components

- [Kubernetes Istio](kubernetesistio)
- [Kubernetes Istio Base CRDs](kubernetesistiobasecrds)
- [Kubernetes Namespace](kubernetesnamespace)
