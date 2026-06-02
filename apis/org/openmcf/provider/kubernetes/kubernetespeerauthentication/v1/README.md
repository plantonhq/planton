# Kubernetes Peer Authentication

Provision an Istio `PeerAuthentication` -- a namespaced policy that sets the
mutual TLS (mTLS) requirements for incoming connections to the workloads it
selects. Use it to require encrypted, authenticated service-to-service traffic
(`STRICT`), to ease a migration onto the mesh (`PERMISSIVE`), or to carve out
specific ports that must stay plaintext.

## What Gets Created

- A namespaced `security.istio.io/v1` `PeerAuthentication` custom resource.
- An optional workload `selector`, a workload-level `mtls.mode`, and optional
  per-port `port_level_mtls` overrides, scoped to this policy's namespace.

## Scope

A PeerAuthentication's reach depends on its selector and namespace:

- **Workload-specific** -- with a `selector`, it applies only to matching pods in
  its namespace.
- **Namespace-wide** -- with no `selector`, it applies to every workload in its
  namespace.
- **Mesh-wide default** -- placed in the Istio root namespace (e.g. `istio-system`)
  with no selector, it becomes the default for the whole mesh.

More specific policies override less specific ones (workload > namespace > mesh).

## Prerequisites

- Istio CRDs installed on the cluster (`KubernetesIstioBaseCrds`).
- A running Istio control plane, istiod (`KubernetesIstio`), to enforce the
  policy. The resource applies with only the CRDs present, but enforcement
  requires istiod and sidecar (or ambient) data-plane proxies.
- The target namespace (`KubernetesNamespace`).

## Quick Start

```yaml
apiVersion: kubernetes.openmcf.org/v1
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
openmcf apply -f peerauthentication.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `namespace` | reference | Namespace the policy is created in (and the scope of a selector-less policy). |

### Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `selector.match_labels` | map | Pod labels selecting the workloads this policy applies to. Omit to cover the whole namespace (or mesh, in the root namespace). |
| `mtls.mode` | string | Workload mTLS mode: `UNSET`, `DISABLE`, `PERMISSIVE`, or `STRICT`. Omit the whole `mtls` block to inherit from the parent policy. |
| `port_level_mtls` | map | Per-port mode overrides keyed by the **workload** port number. Only honored when a `selector` is set. |

### mTLS modes

| Mode | Meaning |
|------|---------|
| `UNSET` | Inherit from the parent policy; if none, behaves as `PERMISSIVE`. |
| `DISABLE` | Plaintext only -- no mTLS tunnel. |
| `PERMISSIVE` | Accept both plaintext and mTLS (useful during migration). |
| `STRICT` | Require an mTLS tunnel (client certificate mandatory). |

## Examples

### Namespace-wide strict mTLS

```yaml
spec:
  namespace:
    value: finance
  mtls:
    mode: STRICT
```

### Strict for one workload, with a plaintext port

```yaml
spec:
  namespace:
    value: finance
  selector:
    match_labels:
      app: finance
  mtls:
    mode: STRICT
  port_level_mtls:
    "8080":
      mode: DISABLE
```

## Composing in Infra Charts

`namespace` is a `StringValueOrRef`, so it can reference a `KubernetesNamespace`
output via `valueFrom`. The `selector.match_labels` is NOT a foreign key -- istiod
matches it against pod labels at runtime, so it creates no automatic DAG edge to
the workload it protects. To order this policy relative to that workload in an
infra chart, declare the dependency on `metadata.relationships`:

```yaml
metadata:
  name: finance-mtls
  relationships:
    - kind: KubernetesDeployment
      name: finance
      type: depends_on
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: finance-ns
      fieldPath: spec.name
  selector:
    match_labels:
      app: finance
  mtls:
    mode: STRICT
```

See `docs/README.md` for the full composability rationale.

## Stack Outputs

| Output | Description |
|--------|-------------|
| `peer_authentication_name` | Name of the created PeerAuthentication (equals metadata.name). |
| `namespace` | Namespace the PeerAuthentication was created in. |

## Related Components

- [Kubernetes Istio](kubernetesistio)
- [Kubernetes Istio Base CRDs](kubernetesistiobasecrds)
- [Kubernetes Namespace](kubernetesnamespace)
