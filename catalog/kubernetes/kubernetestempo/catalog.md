# Grafana Tempo

Deploys Grafana Tempo — cost-efficient distributed tracing — from the official `tempo` Helm chart. Tempo stores whole traces in object storage and needs no expensive index (you retrieve by trace ID or TraceQL), so it scales to high span volumes cheaply. It is the traces half of a complete observability stack alongside metrics (Prometheus) and logs (Loki).

The grain is deliberate: **this kind is the single-binary Tempo** — one StatefulSet, production-capable with an object-storage backend; the `tempo-distributed` microservices chart is deliberately not modeled. Spans arrive over OTLP (from applications or an OpenTelemetry collector at the exported ingest endpoints) and Grafana reads them back (a datasource of type `tempo` at the exported HTTP endpoint). Everything stays ClusterIP; external reachability composes from ingress and Gateway API kinds over the exported Service handle.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Helm release** (official grafana-community `tempo` chart, default pin `2.2.3` pairing with Tempo `2.10.7`, named `metadata.name`) — one Tempo StatefulSet with OTLP receivers on gRPC 4317 and HTTP 4318 always on
- **A PersistentVolumeClaim** (default 10Gi on the cluster's default StorageClass) — the chart's own default is an emptyDir that loses every trace on pod restart; this component deliberately inverts that. With local storage the volume holds ALL trace blocks; with an object-storage backend it holds only the write-ahead log. `ephemeral: true` restores the chart's throwaway posture
- **Object-storage wiring** when a backend is declared — S3/S3-compatible (bucket AND endpoint required; Tempo never derives the endpoint from the region), GCS, or Azure Blob. Empty credentials mean the pod's ambient identity (IRSA on EKS, workload identity on GKE, federated identity on AKS — the recommended keyless postures); declared credentials are references to existing Secrets, injected as environment variables and never rendered into config
- **Kubernetes Namespace** — created only when `createNamespace` is true; otherwise the namespace must already exist

The modules pin the chart's fullname to `metadata.name` so the exported Service and endpoints stay predictable. **Keep the resource name at 45 characters or fewer** — the chart truncates composed child names at 63 characters, so both modules fail loudly on a longer name instead of letting the naming contract break.

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** — an active connection in the Connect module with credentials for the target cluster.

### Kubernetes Cluster

- **A StorageClass** for the data volume (unless `ephemeral`) — most managed clusters provide a default; reference a **Kubernetes StorageClass** for explicit (SSD) placement. Block compaction rewards SSD-backed classes.
- **An object-storage bucket** when a backend is declared — the bucket must exist; Tempo does not create it. An in-cluster **SeaweedFS** S3 endpoint works with `forcePathStyle: true`, and a **Cloudflare R2 bucket** picked from your catalog works through the `r2` arm with nothing S3-shaped to fill in.
- **kube-prometheus-stack** — only if you enable `serviceMonitorEnabled` (the monitoring.coreos.com CRDs) or point the metrics generator at its Prometheus (which must set `prometheus.enableRemoteWriteReceiver: true`).

## Deploy

### Console

Open the deployment store, find **Grafana Tempo**, and click **Deploy**. The creation wizard walks you through namespace placement (with the live naming-budget warning), the chart pin, topology and its volume, the trace-storage backend, retention, the ingest surface, the metrics generator, the query and self-observability dials, sizing, the air-gap path, placement, and the Helm-values escape hatch. Start from the **Production Tempo on object storage** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesTempo
metadata:
  name: traces
  org: acme-corp
  env: prod
spec:
  namespace:
    value: tracing
  createNamespace: true
```

```shell
planton apply -f tempo.yaml
```

This minimal install is a complete trace store: one replica on a 10Gi persistent volume, OTLP receivers on 4317/4318, a 24-hour retention window, single-tenant, and no data leaving the cluster. A Stack Job tracks the provisioning in real time.

### InfraChart

Compose Tempo behind its namespace with a reference, and the InfraPipeline orders the deploys:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: tracing-namespace
      fieldPath: spec.name
  createNamespace: false
```

## Key Configuration

These are the most important decisions when configuring a Tempo installation. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Replicas above 1 require an object-storage backend** — the spec enforces it: replicas cannot share local trace storage. With S3, R2, GCS, or Azure declared, replicas share the backend and scale ingest and query; without one, a single replica is the honest ceiling.

**Persistent by default, ephemeral by choice** — the chart's own emptyDir default loses every trace on pod restart, so this component provisions a PVC unless you opt out. `ephemeral: true` excludes a custom `diskSize` and a `storageClass` (the platform-stamped 10Gi default is tolerated) — there is no volume for them to configure.

**Retention speaks Go durations — there is no day unit** — `retention` accepts minutes or hours only (`30m`, `24h`; a week is `168h`, never `7d`). The chart default of 24h suits a dev loop; raise it for anything users depend on. Longer retention costs volume capacity with local storage and object-store bytes with a backend.

**OTLP-first ingest** — gRPC 4317 and HTTP 4318 are always on; they are the 2026 wire standard. `jaegerReceiversEnabled` opens the four legacy Jaeger protocols (gRPC 14250, thrift-binary 6832, thrift-compact 6831, thrift-http 14268) for fleets still migrating — the component deliberately narrows the chart's all-receivers default, because every closed port is one less ingest surface.

**Multi-tenancy is a header contract** — with `multiTenancyEnabled`, an `X-Scope-OrgID` tenant header is required on every push AND every query; senders and the Grafana datasource must both carry it, or queries return empty results rather than errors.

**The metrics generator lights up Grafana's service map** — it derives service-graph and span metrics from the trace stream and remote-writes them to a Prometheus. The URL accepts a literal or a reference to a **kube-prometheus-stack** (its `prometheus_endpoint` output); the target Prometheus must accept pushes (`prometheus.enableRemoteWriteReceiver: true` on the stack), and when the URL carries no path the modules append the standard `/api/v1/write`. An empty `processors` list runs both `serviceGraphs` and `spanMetrics` — Tempo's own default set.

**Grafana is the query surface** — `tempoQueryEnabled` adds the Jaeger-UI-compatible query sidecar on 16686 only for tooling that speaks the Jaeger API. `usageReporting` is the component's privacy-first divergence from Tempo's report-by-default: no anonymous statistics leave the cluster without an explicit opt-in.

**`helmValues` merges last** — the escape hatch for chart surface beyond the typed fields (per-receiver tuning, tenant overrides, search concurrency). Anything here silently overrides the typed fields on every deploy; never put secrets in it (object-storage credentials belong in the typed Secret-reference fields), and leave `fullnameOverride` alone — the naming contract the outputs derive from depends on it.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|---|---|---|
| Kubernetes Namespace | `spec.namespace` | `spec.name` |
| Kubernetes StorageClass | `spec.storageClass` | `metadata.name` |
| kube-prometheus-stack | `spec.metricsGenerator.remoteWriteUrl` | `status.outputs.prometheus_endpoint` |

Object-storage credentials (`spec.storage.s3.credentials`, `spec.storage.gcs.serviceAccountKeySecret`, `spec.storage.azure.accountKeySecret`) and `spec.imagePullSecrets` name existing Secrets by name + key — references to material that already lives in the cluster, needed only when ambient keyless identity is unavailable.

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | Namespace Tempo runs in | Application deployment manifests |
| `release_name` | Helm release name (= metadata.name); the Service name derives from it | Operational tooling |
| `service` | The Tempo Service (all ports) | Ingress/Gateway exposure |
| `http_endpoint` | In-cluster HTTP endpoint (port 3200) | Grafana `tempo` datasources, TraceQL clients |
| `otlp_grpc_endpoint` | In-cluster OTLP gRPC trace-ingest endpoint (port 4317) | Application SDKs, OpenTelemetry collector `otlp` exporters |
| `otlp_http_endpoint` | In-cluster OTLP HTTP trace-ingest endpoint (port 4318) | OTLP/HTTP senders |
| `port_forward_command` | Copy-paste `kubectl port-forward` for the Tempo API | Local development access |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Dev single-node** — one monolithic replica on a persistent volume with OTLP receivers — the smallest honest trace store; the component defaults carry the whole posture. Start from the **Dev single-node Tempo** preset.

**Production on object storage** — two replicas against an S3-compatible backend (an in-cluster SeaweedFS; AWS S3, GCS or Azure by swapping the block), a two-week retention window, and the metrics generator remote-writing to the cluster's Prometheus. Start from the **Production Tempo on object storage** preset.

**Jaeger migration bridge** — the four legacy Jaeger receiver protocols opened alongside OTLP and the Jaeger-UI-compatible query sidecar — a bridge posture; disable the receivers once emitters are on OTLP. Start from the **Jaeger-compatible Tempo** preset.

## Works With

- [**Kubernetes Namespace**](/cloud-catalog/kubernetes-namespace) — referenced placement; the InfraPipeline orders namespace-first.
- [**Kubernetes StorageClass**](/cloud-catalog/kubernetes-storage-class) — SSD-backed classes for the data volume.
- [**Kubernetes Secret**](/cloud-catalog/kubernetes-secret) — object-storage credentials and image-pull Secrets, always by reference.
- [**SeaweedFS**](/cloud-catalog/kubernetes-seaweed-fs) — the in-cluster S3-compatible backend for the storage block.
- [**OpenTelemetry Collector**](/cloud-catalog/kubernetes-otel-collector) — sends spans to the exported OTLP endpoints from a traces pipeline.
- [**kube-prometheus-stack**](/cloud-catalog/kubernetes-kube-prometheus-stack) — receives the metrics generator's remote-write (the service map's data) and provides the ServiceMonitor CRDs.
- [**Grafana**](/cloud-catalog/kubernetes-grafana) — reads traces back through a `tempo` datasource at the exported HTTP endpoint.
- [**Kubernetes Ingress**](/cloud-catalog/kubernetes-ingress) — HTTP exposure over the exported Service handle (Gateway API kinds compose the same way).
