# Kubernetes Qdrant

## When NOT to Use This

**One resource is ONE Qdrant cluster.** The chart's grain is a
StatefulSet whose size is the `replicas` field: distributed mode is
always on (the chart's default), pod 0 bootstraps the Raft consensus,
and every further replica joins over the p2p port. Scaling the
deployment is a `replicas` change — nothing else.

Also not the right kind when:

- **You want shards or replication factor in the manifest** — those
  are collection-level properties declared per collection through the
  Qdrant API (DATA operations), not deployment configuration. Adding
  pods alone does not copy data.
- **You want a managed vector database** — this component runs the
  open-source Qdrant (Apache-2.0) ON the Kubernetes cluster itself,
  from the official chart.
- **You expect a public endpoint out of the box** — the Service is
  ClusterIP; exposure composes from first-class kinds
  (KubernetesIngress, Gateway API kinds) over the service handle.
- **You expect authentication out of the box** — API keys are OFF
  upstream and off by default here; declare `api_key` for anything
  beyond a private dev cluster.

## Overview

**KubernetesQdrant** deploys Qdrant — the catalog's vector database
(Apache-2.0), the search engine behind RAG, semantic search,
agent-memory and recommendation architectures — from the official
`qdrant` Helm chart (https://qdrant.github.io/qdrant-helm). Vector
databases store embeddings and answer nearest-neighbor queries over
them; Qdrant pairs that with payload filtering and quantization,
which is what retrieval layers depend on at scale.

**The API-key contract**: keys never land in rendered Helm values.
`api_key` (read-write) and `read_only_api_key` each take one of two
arms — `generate: true` makes the chart generate a random key ONCE at
first install (stable across upgrades), while `existing_secret`
points the chart at a Secret you own (it must exist BEFORE the
install; the chart reads it at template time). Either way the key
material lives in the chart-owned `<name>-apikey` Secret (keys
`api-key` / `read-only-api-key`), whose name lands in the stack
outputs. A read-only key REQUIRES a read-write key — the spec
enforces it (an unauthenticated cluster with a read-only key protects
nothing).

**Key design points:**

- **Distributed mode is always on.** The chart ships
  `config.cluster.enabled: true`; a single replica is simply a
  one-member cluster. REST is 6333, gRPC 6334 (SDKs default to gRPC),
  p2p is 6335. The Service name equals the resource name
  (fullnameOverride is pinned to `metadata.name`), which is what
  makes the exported endpoints deterministic.
- **TLS is a seam.** `tls` mounts an existing TLS Secret
  (`tls.crt`/`tls.key` — the shape cert-manager produces) at
  `/qdrant/tls` and enables HTTPS/gRPC-TLS on the client listeners;
  the chart's probes switch scheme automatically. Inter-node p2p TLS
  is a separate upstream surface (`config.cluster.p2p`) and rides
  `helm_values`.
- **Storage is per pod.** The `storage` PVC (default 10Gi) carries
  vectors, payloads and the write-ahead log; declare the separate
  `snapshots` volume when using snapshots, sized like the main volume
  (upstream's guidance) so a large snapshot cannot crash the node.
  Size container MEMORY for the vectors you plan to hold — Qdrant
  keeps hot segments and indexes in RAM.
- **`helm_values` is the escape hatch** — additional chart values
  merged LAST over everything the typed fields render (Helm `-f`
  semantics, identical on both engines); for the chart surface beyond
  the typed fields (the `config:` engine document — collection
  defaults, optimizer/WAL tuning, p2p TLS — probes, PDB, sidecars,
  additional volumes), never a substitute for them. Never for
  secrets.

## Essential Configuration Fields

### Required

- **`spec.namespace`**: namespace to install into — literal or a
  KubernetesNamespace reference (`create_namespace` to own it)

### Common

- **`spec.chart_version`**: chart pin (default `1.18.2` — the chart
  version tracks the Qdrant release it ships)
- **`spec.replicas`**: cluster size (default 1); any higher count
  forms a Raft cluster over the p2p port
- **`spec.resources`**: container resources — size memory for the
  vectors you plan to hold
- **`spec.storage` / `spec.snapshots`**: per-pod PVCs — size (default
  10Gi) and StorageClass (literal or a KubernetesStorageClass
  reference; empty = the cluster default); declare `snapshots` when
  using snapshots, sized like `storage`
- **`spec.api_key` / `spec.read_only_api_key`**: `generate: true` or
  `existing_secret {name, key}`; read-only requires read-write
- **`spec.tls`**: existing TLS Secret (`tls.crt`/`tls.key`) for
  HTTPS/gRPC-TLS on the client listeners; empty = plaintext
  in-cluster
- **`spec.scheduling`**: node selector, tolerations,
  `pod_anti_affinity` (required anti-affinity across hostnames —
  meaningful from 2 replicas up), priority class
- **`spec.service_monitor_enabled`**: Prometheus ServiceMonitor for
  `/metrics` (requires the Prometheus Operator CRDs)
- **`spec.image` / `spec.helm_values`**: the air-gap image path
  (repository, tag, `use_unprivileged` for restricted Pod Security
  Standards) and the escape hatch

## Outputs

| Output | Purpose |
|---|---|
| `namespace` | Namespace the cluster runs in |
| `release_name` | Helm release name (= `metadata.name`) |
| `service_name` | The main Qdrant Service (http 6333, grpc 6334; = the release name) |
| `http_endpoint` | In-cluster REST endpoint (`http(s)://<name>.<ns>.svc.cluster.local:6333`) |
| `grpc_endpoint` | In-cluster gRPC endpoint SDKs default to (`<name>.<ns>.svc.cluster.local:6334`) |
| `api_key_secret_name` | The chart-owned `<name>-apikey` Secret (key `api-key`); empty when unauthenticated |
| `read_only_api_key_secret_name` | The same Secret (key `read-only-api-key`); empty when not declared |
| `port_forward_command` | Port-forward command for REST access from a workstation |

## Composing in Infra Charts

- **`spec.namespace`** is a foreign key (default kind
  KubernetesNamespace); the **storage/snapshots `storage_class`**
  references a KubernetesStorageClass; **`tls.secret`** accepts a
  KubernetesCertificate reference (its issued Secret already carries
  `tls.crt`/`tls.key` — the shape the module mounts).
- **Applications consume the outputs**: `grpc_endpoint` as the SDK
  URI (or `http_endpoint` for REST), the read-write key from
  `api_key_secret_name` for ingestion services, the read-only key
  from `read_only_api_key_secret_name` for query-only consumers —
  the key rides the Secret, never the manifest.
- **Exposure composes, never embeds**: a KubernetesIngress or Gateway
  API route over `service_name` for REST; gRPC needs an HTTP/2-aware
  route.

## Examples

The smallest declarable Qdrant is a namespace alone — every other
field has a working default (single node, no auth, 10Gi storage).
The production shape:

### Production cluster (3 nodes, generated keys, sized)

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesQdrant
metadata:
  name: vectors
spec:
  namespace:
    value: vector
  create_namespace: true
  replicas: 3
  api_key:
    generate: true
  read_only_api_key:
    generate: true
  resources:
    requests: { cpu: "1", memory: 8Gi }
    limits: { cpu: "2", memory: 8Gi }
  storage:
    size: 50Gi
  snapshots:
    size: 50Gi
  scheduling:
    pod_anti_affinity: true
  service_monitor_enabled: true
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
