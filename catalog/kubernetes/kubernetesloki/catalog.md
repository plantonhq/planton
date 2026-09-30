# Grafana Loki

Deploys Grafana Loki — horizontally scalable, cost-efficient log aggregation — from the official `loki` Helm chart. Loki indexes only the *labels* of your logs and stores the compressed content in object storage, so it stays cheap at high volume: the logs half of a complete observability stack alongside metrics (Prometheus) and traces (Tempo).

The grain is deliberate: **Loki stores logs; something must ship them.** Deploy a KubernetesOtelCollector in daemonset mode with the cluster-logs pipeline (its presets carry it) pointed at this component's exported `gateway_endpoint` — and read them back by pointing a KubernetesGrafana datasource of type `loki` at the same endpoint. Everything stays ClusterIP; the nginx gateway is the one front door.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Helm release** (official `loki` chart, default pin `18.5.4` — pairs with Loki 3.7.4, named `metadata.name`) — a single-replica monolithic StatefulSet on a persistent volume by default, or the topology the spec declares; plus the nginx **gateway** (the single front door routing pushes and queries in every mode), the **chunks and query-results memcached caches**, and the **canary** DaemonSet that writes and reads test log lines through the full pipeline, turning silent log loss into a visible metric
- **A derived index schema** — Loki normally requires a hand-authored `schema_config`; the modules derive it (TSDB, schema v13, the object store matching your storage backend) so a new install never writes one. The `schemaFromDate` override exists solely for importing clusters whose existing schema began on a real date
- **Kubernetes Namespace** — created only when `createNamespace` is true; otherwise the namespace must already exist

The modules pin the chart's fullname to `metadata.name` so child names stay predictable (`<name>-gateway`, `<name>-backend-headless`). **Keep the resource name at 40 characters or fewer** — the chart truncates composed child names at 63 characters, so the modules fail loudly over the budget instead of letting the naming contract break.

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** — an active connection in the Connect module with credentials for the target cluster.

### Kubernetes Cluster

- **An object-storage bucket** for anything beyond a single filesystem replica — S3 (or an in-cluster **Kubernetes SeaweedFS** S3 endpoint), a **Cloudflare R2 bucket** picked from your catalog, GCS, or Azure Blob. The bucket must exist; Loki does not create it.
- **A StorageClass** for the persistent volumes — most managed clusters provide a default.
- **kube-prometheus-stack** — only if you set `serviceMonitorEnabled` (the monitoring.coreos.com CRDs) or wire the ruler's `alertmanagerUrl` by reference.

## Deploy

### Console

Open the deployment store, find **Grafana Loki**, and click **Deploy**. The creation wizard walks you through namespace placement (with the live naming-budget warning), the chart pin, the deployment mode, object storage, retention and the import-only schema override, the ingestion/query limits, multi-tenancy, the gateway, the caches (with the live sizing warning), the ruler, self-observability, the air-gap path, placement, and the Helm-values escape hatch. Start from the **Dev single-node Loki** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesLoki
metadata:
  name: central-logs
  org: acme-corp
  env: prod
spec:
  namespace:
    value: logging
  createNamespace: true
  caching:
    chunksCacheMemoryMb: 256
    resultsCacheMemoryMb: 128
```

```shell
planton apply -f loki.yaml
```

This near-empty spec is a complete log store: a single-replica monolithic Loki on a 10Gi persistent volume, the gateway, both caches (sized down here for a small cluster — see below), the canary, and single-tenant access with no tenant header needed. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to compose Loki behind a namespace managed by another Cloud Resource:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: logging-namespace
      fieldPath: spec.name
  createNamespace: false
```

The InfraPipeline deploys the namespace first, then provisions Loki into it.

## Key Configuration

These are the most important decisions when configuring a Loki log store. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Two honest topologies — and the storage doctrine that binds them.** `monolithic` (the default when nothing is declared) runs every Loki target in one StatefulSet: right for single-node clusters, dev environments and small production volumes. `simpleScalable` splits into write/read/backend tiers that scale independently and REQUIRES object storage — the tiers rendezvous in the object store, not on a shared disk. Filesystem storage is honest ONLY for a single monolithic replica; more than one replica, or any scalable tier, requires s3/r2/gcs/azure (mirroring the chart's own validation). The chart's microservices mode is deliberately not modeled.

**Object storage is keyless first.** On EKS leave the S3 credentials empty for ambient IRSA identity; on GKE the missing key Secret means workload identity; on AKS the missing account key means federated identity. Declared credentials are name+key REFERENCES to existing Secrets — the modules inject them as environment variables (S3) or a mounted key file (GCS); they never land in the rendered Loki config. The s3-compatible arm (`endpoint` + `forcePathStyle`) composes with an in-cluster KubernetesSeaweedFs. The chart's bundled MinIO subchart is deprecated by the chart itself and is never enabled by this component.

**THE CACHE-SIZING TRAP (verified live):** the caching defaults are production-scale. Unset, the chunks cache allocates 8192MB and the chart requests container memory at 1.2× — a **9830Mi request that never schedules** on a node with less than ~10Gi allocatable. The pod stays Pending and, because the install is atomic, the WHOLE release rolls back after its timeout. Set `chunksCacheMemoryMb` explicitly on any small or dev cluster (128–1024 is plenty for light query loads; the dev preset carries 256).

**Retention is off by default — deliberately visible.** Empty `retentionPeriod` keeps everything forever (Loki's own default) and object-storage costs grow unbounded; production installs should set it (`744h` / `31d` — hours or days only). Deletion is asynchronous: the compactor marks and later sweeps, so any bucket lifecycle policy must expire LATER than the period, never earlier.

**Single-tenant by default — one line of wiring.** This component diverges from the chart's multi-tenant-on default so pushes and queries need no `X-Scope-OrgID` header. Enable `multiTenancy` for isolation: every client then sends its tenant header, and the gateway enforces HTTP basic auth for the declared tenants (name + bcrypt htpasswd hash — one-way material, safe in a manifest; generate with `htpasswd -nbBC10`) or an existing htpasswd Secret. Basic auth AUTHENTICATES clients; each client still declares its own tenant header.

**The gateway is the one front door.** It routes pushes and queries to the right internal target in every mode, and the exported endpoints assume it — disable it only when clients address the internal services directly (single-tenant monolithic only). Expose it via KubernetesIngress or Gateway API kinds over the exported handles; Loki never opens its own doors.

**Alerting on logs** — the ruler evaluates LogQL alerting/recording rules discovered from ConfigMaps labeled `loki_rule: "1"` (the same sidecar contract Grafana dashboards use) and fires at `alertmanagerUrl` — a literal URL or a KubernetesKubePrometheusStack reference (its Alertmanager endpoint), the one-line wiring into the cluster's alerting.

**`helmValues` merges last** — the escape hatch for chart surface beyond the typed fields (bloom filters, the pattern ingester, zone-aware rollouts, per-component overrides, ruler storage tuning). Anything here silently overrides the typed fields on every deploy; never put secrets in it, and leave `fullnameOverride` alone — the naming contract the outputs derive from depends on it.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |
| **KubernetesStorageClass** | `monolithic.storageClass` / `simpleScalable.storageClass` | `metadata.name` |
| **KubernetesKubePrometheusStack** | `ruler.alertmanagerUrl` | `status.outputs.alertmanager_endpoint` |

Object-store credentials (`storage.s3.credentials`, `storage.gcs.serviceAccountKeySecret`, `storage.azure.accountKeySecret`), the bring-your-own htpasswd Secret (`multiTenancy.existingHtpasswdSecret`), and `imagePullSecrets` are name+key references to EXISTING Secrets in the install namespace — not foreign keys; declare them only when ambient keyless identity (IRSA / GKE workload identity / AKS federated identity) is unavailable.

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | Namespace Loki runs in | Application deployment manifests |
| `release_name` | Helm release name (= metadata.name); every child name derives from it | Operational tooling |
| `gateway_service` | The gateway Service (`<name>-gateway`, port 80). Empty when the gateway is disabled | Ingress/Gateway exposure |
| `gateway_endpoint` | In-cluster gateway URL — log shippers push here and Grafana `loki` datasources read here | Collector pipelines, Grafana datasources |
| `otlp_push_endpoint` | The gateway's `/otlp` route — point a KubernetesOtelCollector otlphttp exporter here | OTLP log ingest |
| `loki_service` | The Loki HTTP Service (`<name>`, port 3100) — the direct internal API behind the gateway | Diagnostics, direct API clients |
| `port_forward_command` | Copy-paste `kubectl port-forward` for the gateway | Local development access |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Dev loop / small single-node cluster** — the smallest honest Loki: one monolithic replica on a filesystem volume with the gateway — and the caches sized down explicitly (256/128MB), the one knob a small cluster cannot leave to defaults. Start from the **Dev single-node Loki** preset.

**Production log volume** — write/read/backend tiers (3/3/3) on an in-cluster SeaweedFS S3 bucket with credential references, 31-day compactor-enforced retention, and raised ingest limits. Start from the **Production scalable Loki** preset.

**Shared logging service** — the scalable shape plus tenant isolation: every push/query carries `X-Scope-OrgID`, and the gateway enforces per-team basic auth from bcrypt htpasswd hashes — no Secret fixture needed. Start from the **Multi-tenant shared Loki** preset.

## Works With

- [**OpenTelemetry Collector**](/cloud-catalog/kubernetes-otel-collector) — ships logs in: daemonset mode with the cluster-logs pipeline pointed at the exported `gateway_endpoint` / `otlp_push_endpoint`
- [**Grafana**](/cloud-catalog/kubernetes-grafana) — reads logs back: a datasource of type `loki` at the gateway endpoint
- [**kube-prometheus-stack**](/cloud-catalog/kubernetes-kube-prometheus-stack) — scrapes Loki's ServiceMonitors and receives the ruler's log alerts by reference
- [**SeaweedFS**](/cloud-catalog/kubernetes-seaweed-fs) — in-cluster S3-compatible object storage for the s3 arm (`endpoint` + `forcePathStyle`)
- [**Kubernetes Namespace**](/cloud-catalog/kubernetes-namespace) — referenced placement; the InfraPipeline orders namespace-first
- [**Kubernetes StorageClass**](/cloud-catalog/kubernetes-storage-class) — explicit volume classes for the data/WAL volumes
- [**Kubernetes Ingress**](/cloud-catalog/kubernetes-ingress) — HTTP exposure over the exported gateway Service handle
- [**Kubernetes Gateway**](/cloud-catalog/kubernetes-gateway) — the Gateway API alternative for exposing the gateway Service
