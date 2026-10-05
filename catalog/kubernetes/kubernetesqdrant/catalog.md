# Qdrant

Deploys a Qdrant vector database — the engine behind semantic search, RAG retrieval and recommendation workloads — on any Kubernetes cluster from the official `qdrant` Helm chart (qdrant.github.io/qdrant-helm). Distributed mode is always on: every install is a Raft cluster, even at one member, so growing later is only a `replicas` change — the new pods join the existing consensus, no migration. Collections replicate per their OWN `replication_factor` (a data operation through the Qdrant API); pods make room for replicas, they do not create them.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Kubernetes Namespace** — created (with the standard Planton governance labels) only when `createNamespace` is `true`; otherwise the namespace must already exist
- **Qdrant Helm Release** — the official chart at the pinned `chartVersion` (default 1.18.2), which creates:
  - a StatefulSet of `replicas` pods (default 1) with your CPU/memory resources; pod 0 bootstraps Raft consensus, every further pod joins over the p2p port (6335)
  - a **ClusterIP Service** carrying REST 6333 and gRPC 6334 — what in-cluster clients use, and what the exported endpoints point at; exposure beyond the cluster composes from first-class kinds (KubernetesIngress, Gateway API) over the exported service handle
  - a **PersistentVolumeClaim per pod** for the data volume (10Gi on the cluster's default StorageClass unless configured) — vectors, payloads, and the write-ahead log
  - a **separate snapshots PVC per pod** — only when the `snapshots` block is declared; otherwise snapshots land on the data volume and can fill it
- **API-key Secret** — when a key uses the `generate` arm, the chart mints it ONCE at first install (stable across upgrades) and keeps it in the chart-owned Secret; key material never enters a manifest, and the Secret name is exported in the outputs
- **ServiceMonitor** — only when `serviceMonitorEnabled` is `true` (requires the Prometheus Operator CRDs on the cluster)

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** — an active connection in the Connect module with kubeconfig credentials for the target Kubernetes cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** — required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **A StorageClass** capable of dynamic PV provisioning for the per-pod volumes. Fast (SSD-backed) storage pays off on cold starts and segment optimization.
- **Prometheus Operator CRDs** — only when enabling the ServiceMonitor; without them the install fails on the unknown resource type.
- **An existing API-key Secret** — only when a key uses the `existingSecret` arm; the chart resolves it at template time, so it must exist before the first install.
- **An existing TLS Secret** — only when declaring the `tls` block; a standard `tls.crt`/`tls.key` Secret (the shape cert-manager produces).

## Deploy

### Console

Open the deployment store, find **Qdrant**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Dev single node preset** in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesQdrant
metadata:
  name: prod-qdrant
  org: acme-corp
  env: prod
spec:
  namespace:
    value: qdrant
  createNamespace: true
  replicas: 3
  apiKey:
    generate: true
  readOnlyApiKey:
    generate: true
  resources:
    requests:
      cpu: "1"
      memory: 8Gi
    limits:
      cpu: "2"
      memory: 8Gi
  storage:
    size: 50Gi
  snapshots:
    size: 50Gi
  scheduling:
    podAntiAffinity: true
```

```shell
planton apply -f qdrant.yaml
```

This creates a production-shaped 3-node cluster: the quorum posture (Raft survives one member loss), generated read-write and read-only keys living in the chart-owned Secret, 8Gi of memory per node for the hot vector working set, equally-sized data and snapshots volumes per pod, and required anti-affinity so a node loss takes one member — not the quorum.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the cluster to a namespace managed by another Infra Component:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: qdrant-namespace
      fieldPath: spec.name
  createNamespace: false
```

The InfraPipeline deploys the namespace first, then provisions the Qdrant cluster into it.

## Key Configuration

These are the most important decisions when configuring Qdrant. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Cluster size** — `replicas` defaults to 1, which is still a one-member Raft cluster (distributed mode is always on). From 2 up, remember where deployment stops and data begins: collections replicate per their own `replication_factor`, set through the Qdrant API when the collection is created — adding pods alone does not copy data. Memory is the real capacity bound: Qdrant serves from RAM-resident segments and HNSW indexes, so size `resources` for the vectors each node will hold (vectors × dimensions × 4 bytes plus index overhead is the back-of-envelope; quantization, a collection-level setting, stretches it severalfold).

**API keys** — `apiKey` (read-write) and `readOnlyApiKey` each take at most one arm: `generate` (the chart mints the key once at first install and keeps it in the chart-owned Secret — key material never enters a manifest) or `existingSecret` (name + key coordinates of a Secret that exists before the install). Empty means **unauthenticated** — the upstream default; the listeners accept any request, acceptable only on a private dev cluster inside a namespace boundary. A read-only key requires a read-write key (the spec enforces the pair): with no read-write key the listeners are open anyway, so a read-only key would protect nothing. Hand the read-only key to query-side consumers so a leaked retrieval credential cannot mutate collections.

**Storage and snapshots** — `storage.size` defaults to 10Gi PER POD on the cluster's default StorageClass; `storage.storageClass` accepts a literal class name or a KubernetesStorageClass reference. The separate `snapshots` block is upstream's recommendation for anything real: snapshots (backups, snapshot shard transfers) written onto the data volume can fill it and crash the node. Size it like `storage` — a snapshot is roughly a copy of the data. Growing a PVC later depends on the class allowing expansion.

**TLS** — the `tls` block has one field: `secret`, an existing TLS Secret (`tls.crt`/`tls.key`, the shape cert-manager produces; a KubernetesCertificate reference resolves to its Secret name). Both client listeners (REST 6333, gRPC 6334) switch to TLS together. Empty means plaintext in-cluster — compose TLS at the exposure layer instead. Inter-node p2p TLS is a separate upstream surface (`config.cluster.p2p`) that rides `helmValues`.

**Scheduling** — `scheduling.podAntiAffinity` renders REQUIRED anti-affinity across hostnames: every member on a different node, so one node loss takes one member instead of the quorum. Chart default is none; from 2 replicas up this is the production posture — but required means unschedulable pods when nodes run short, so keep node count ≥ replicas. `nodeSelector`, `tolerations` and `priorityClassName` carry the standard placement dials.

**Image** — `image.repository` (INCLUDING the registry) and `image.tag` override the official `qdrant/qdrant` for air-gapped mirrors. `image.useUnprivileged` appends `-unprivileged` to the effective tag and skips the root-owned volume-ownership init container — the restricted Pod Security Standards path (a mirror must carry that tag variant too).

**Escape hatch** — `helmValues` merges LAST over everything the typed fields render (Helm `-f` semantics, identical on both engines). The chart renders `config:` verbatim into the engine's production.yaml — engine tuning (request-size ceilings, collection defaults, optimizer settings, p2p TLS) rides here. Never a substitute for the typed fields, and never a place for secrets; key material belongs in the `apiKey` arms.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |
| **KubernetesStorageClass** | `storage.storageClass`, `snapshots.storageClass` | `status.outputs.storage_class_name` |
| **KubernetesCertificate** | `tls.secret` | `status.outputs.secret_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | Namespace the cluster runs in | Application deployment manifests |
| `release_name` | Helm release name (= metadata.name) | Operational tooling |
| `service_name` | Name of the Qdrant Service (REST/gRPC ports) | Ingress/Gateway backends, NetworkPolicies |
| `http_endpoint` | In-cluster REST endpoint (port 6333) | Dashboard access and REST clients |
| `grpc_endpoint` | In-cluster gRPC endpoint (port 6334) — what SDKs default to | Application client connection strings |
| `api_key_secret_name` | Secret holding the read-write API key (chart-owned when generated, or the referenced Secret); empty when unauthenticated | Workload env wiring — read the key from the Secret, never an env literal |
| `read_only_api_key_secret_name` | Secret holding the read-only API key; empty when not declared | Query-side workload env wiring |
| `port_forward_command` | Command to port-forward the listeners to a developer laptop | Local development |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Dev single node** — the smallest declarable Qdrant: one node, no authentication (the upstream default — private dev clusters only), a 5Gi volume, the chart's own resource defaults. Start from the **Dev single node preset**.

**Production cluster** — 3 nodes (the quorum posture), generated read-write and read-only keys, 8Gi of memory per node, equally-sized 50Gi data and snapshots volumes on fast storage, required anti-affinity, a ServiceMonitor. Start from the **Production cluster preset**.

**RAG workload** — a single node sized for a typical RAG corpus (8Gi of memory holds an embedding set in the several-million-vector range at common dimensions), a generated key because ingestion and retrieval cross service boundaries, a snapshots volume so the index restores in minutes instead of re-embedding for hours, and a `helmValues` entry raising the request-size ceiling for bulk-ingestion batches. Start from the **RAG workload preset**.

## Works With

- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) — provides the namespace for the cluster
- [**Kubernetes StorageClass**](/infra-catalog/kubernetes-storage-class) — fast storage for the data volume, cold storage for snapshots
- [**Cert Manager Certificate**](/infra-catalog/kubernetes-certificate) — cert-manager-issued TLS Secrets for the client listeners
- [**Kubernetes Ingress**](/infra-catalog/kubernetes-ingress) — composes exposure over the exported service handle instead of a direct LoadBalancer
