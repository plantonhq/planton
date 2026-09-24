# KubernetesTempo

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

**KubernetesTempoSpec** deploys Grafana Tempo — the distributed-tracing
backend that stores whole traces in object storage and retrieves them
by ID or TraceQL — from the official monolithic `tempo` Helm chart
(https://grafana-community.github.io/helm-charts).

HOW TRACES GET IN: applications (or a KubernetesOtelCollector in
between) send OTLP to the exported `otlp_grpc_endpoint` /
`otlp_http_endpoint`. Grafana reads them back: point a
KubernetesGrafana datasource of type `tempo` at the exported
`http_endpoint`.

GRAIN: this kind models the SINGLE-BINARY Tempo — one StatefulSet,
production-capable with an object-storage backend, the right grain
for the vast majority of tracing volumes. The separate
`tempo-distributed` microservices chart is deliberately not modeled.

STORAGE: `local` (the default) keeps trace blocks on a
PersistentVolume — honest for a single replica. More than one
replica REQUIRES an object-storage backend (s3/gcs/azure); the
s3-compatible arm composes with an in-cluster KubernetesSeaweedFs.
KNOW THIS: the chart's own default runs on an emptyDir — every trace
vanishes on pod restart — so this component provisions a
PersistentVolumeClaim by default instead (`ephemeral` restores the
chart posture).

METRICS FROM TRACES: the metrics generator derives service-graph and
span-duration metrics from the trace stream and remote-writes them to
a Prometheus — the seam that lights up Grafana's service map. Point
it at a KubernetesKubePrometheusStack (whose Prometheus must set
`enable_remote_write_receiver: true`).

EXPOSURE: everything stays ClusterIP; expose via first-class kinds
(KubernetesIngress, Gateway API kinds) over the exported handles.
Keep the resource name at 45 characters or fewer — the chart
truncates composed child names at 63 characters; the modules pin the
chart's fullname to the resource name and fail loudly over the
budget.

The typed fields cover the chart's meaningful surface; `helm_values`
remains the escape hatch (merged last, Helm `-f` semantics, identical
on both engines) for the long tail — per-receiver tuning, tenant
overrides, search concurrency — a safety valve, never the primary
interface.

## Example

```yaml
# Full-surface offline-proof manifest: multiple replicas on an
# S3-compatible object store (credentials via Secret + env expansion), a
# longer retention, the legacy Jaeger receivers, multi-tenancy, the
# trace-derived metrics generator remote-writing to a Prometheus, the
# tempo-query sidecar, ServiceMonitor, a private-mirror image registry with
# a pull secret, and scheduling — so the offline tofu plan and pulumi
# preview proofs cover the full typed surface. Placeholder values; never
# applied to a real cluster.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesTempo
metadata:
  name: tempo-hack
spec:
  namespace:
    value: tempo-hack
  createNamespace: true
  chartVersion: 2.2.3
  replicas: 2
  storage:
    s3:
      bucket: tempo-traces
      endpoint: objects-filer.storage.svc.cluster.local:8333
      forcePathStyle: true
      insecure: true
      credentials:
        accessKeyIdSecret:
          name: tempo-s3
          key: access-key-id
        secretAccessKeySecret:
          name: tempo-s3
          key: secret-access-key
  retention: 336h
  jaegerReceiversEnabled: true
  multiTenancyEnabled: true
  metricsGenerator:
    enabled: true
    remoteWriteUrl:
      value: http://monitoring-kube-prometheus-prometheus.observability.svc.cluster.local:9090
    processors:
      - service_graphs
      - span_metrics
  resources:
    requests:
      cpu: 100m
      memory: 256Mi
    limits:
      cpu: "1"
      memory: 2Gi
  tempoQueryEnabled: true
  serviceMonitorEnabled: true
  imageRegistry: mirror.example.com
  imagePullSecrets:
    - mirror-pull
  scheduling:
    nodeSelector:
      kubernetes.io/os: linux
    tolerations:
      - key: dedicated
        operator: Equal
        value: observability
        effect: NoSchedule
    priorityClassName: system-cluster-critical
  helmValues: |
    tempo:
      podLabels:
        team: observability
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.createNamespace` | `bool` |  |  |  |
| `spec.chartVersion` | `string` |  | `2.2.3` |  |
| `spec.replicas` | `int32` |  | `1` |  |
| `spec.storage` | `KubernetesTempoStorage` |  |  |  |
| `spec.storage.local` | `KubernetesTempoLocalStorage` |  |  |  |
| `spec.storage.s3` | `KubernetesTempoS3Storage` |  |  |  |
| `spec.storage.s3.bucket` | `string` | yes |  |  |
| `spec.storage.s3.endpoint` | `string` | yes |  |  |
| `spec.storage.s3.region` | `string` |  |  |  |
| `spec.storage.s3.forcePathStyle` | `bool` |  |  |  |
| `spec.storage.s3.insecure` | `bool` |  |  |  |
| `spec.storage.s3.credentials` | `KubernetesTempoObjectStoreCredentials` |  |  |  |
| `spec.storage.s3.credentials.accessKeyIdSecret` | `KubernetesTempoSecretKeyRef` | yes |  |  |
| `spec.storage.s3.credentials.accessKeyIdSecret.name` | `string` | yes |  |  |
| `spec.storage.s3.credentials.accessKeyIdSecret.key` | `string` | yes |  |  |
| `spec.storage.s3.credentials.secretAccessKeySecret` | `KubernetesTempoSecretKeyRef` | yes |  |  |
| `spec.storage.s3.credentials.secretAccessKeySecret.name` | `string` | yes |  |  |
| `spec.storage.s3.credentials.secretAccessKeySecret.key` | `string` | yes |  |  |
| `spec.storage.gcs` | `KubernetesTempoGcsStorage` |  |  |  |
| `spec.storage.gcs.bucket` | `string` | yes |  |  |
| `spec.storage.gcs.serviceAccountKeySecret` | `KubernetesTempoSecretKeyRef` |  |  |  |
| `spec.storage.gcs.serviceAccountKeySecret.name` | `string` | yes |  |  |
| `spec.storage.gcs.serviceAccountKeySecret.key` | `string` | yes |  |  |
| `spec.storage.azure` | `KubernetesTempoAzureStorage` |  |  |  |
| `spec.storage.azure.accountName` | `string` | yes |  |  |
| `spec.storage.azure.container` | `string` | yes |  |  |
| `spec.storage.azure.accountKeySecret` | `KubernetesTempoSecretKeyRef` |  |  |  |
| `spec.storage.azure.accountKeySecret.name` | `string` | yes |  |  |
| `spec.storage.azure.accountKeySecret.key` | `string` | yes |  |  |
| `spec.diskSize` | `string` |  | `10Gi` |  |
| `spec.storageClass` | `string \| valueFrom` |  |  | KubernetesStorageClass (`status.outputs.storage_class_name`) |
| `spec.ephemeral` | `bool` |  |  |  |
| `spec.retention` | `string` |  | `24h` |  |
| `spec.jaegerReceiversEnabled` | `bool` |  | `false` |  |
| `spec.multiTenancyEnabled` | `bool` |  |  |  |
| `spec.metricsGenerator` | `KubernetesTempoMetricsGenerator` |  |  |  |
| `spec.metricsGenerator.enabled` | `bool` |  |  |  |
| `spec.metricsGenerator.remoteWriteUrl` | `string \| valueFrom` |  |  | KubernetesKubePrometheusStack (`status.outputs.prometheus_endpoint`) |
| `spec.metricsGenerator.processors` | `[]enum` |  |  |  |
| `spec.resources` | `ContainerResources` |  |  |  |
| `spec.resources.limits` | `CpuMemory` |  |  |  |
| `spec.resources.limits.cpu` | `string` |  |  |  |
| `spec.resources.limits.memory` | `string` |  |  |  |
| `spec.resources.requests` | `CpuMemory` |  |  |  |
| `spec.resources.requests.cpu` | `string` |  |  |  |
| `spec.resources.requests.memory` | `string` |  |  |  |
| `spec.tempoQueryEnabled` | `bool` |  |  |  |
| `spec.serviceMonitorEnabled` | `bool` |  |  |  |
| `spec.usageReporting` | `bool` |  | `false` |  |
| `spec.imageRegistry` | `string` |  |  |  |
| `spec.imagePullSecrets` | `[]string` |  |  |  |
| `spec.scheduling` | `KubernetesTempoScheduling` |  |  |  |
| `spec.scheduling.nodeSelector` | `map<string, string>` |  |  |  |
| `spec.scheduling.tolerations` | `[]WorkloadToleration` |  |  |  |
| `spec.scheduling.tolerations[].key` | `string` |  |  |  |
| `spec.scheduling.tolerations[].operator` | `string` |  |  |  |
| `spec.scheduling.tolerations[].value` | `string` |  |  |  |
| `spec.scheduling.tolerations[].effect` | `string` |  |  |  |
| `spec.scheduling.tolerations[].tolerationSeconds` | `int64` |  |  |  |
| `spec.scheduling.priorityClassName` | `string` |  |  |  |
| `spec.helmValues` | `string` |  |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

Namespace to install into. Accepts a literal namespace name or a
reference to a KubernetesNamespace resource.

- references: KubernetesNamespace (`spec.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.createNamespace

`bool`

When true, the namespace is created (with the standard Planton
governance labels) before installing and deleted with the resource.
When false, the namespace must already exist.

### spec.chartVersion

`string` · optional (explicit presence)

Helm chart version to install (e.g. "2.2.3" — chart 2.2.3 pairs
with Tempo 2.10.7). Versions must exist as SERVED charts in the
repository index (https://grafana-community.github.io/helm-charts).

- default: `2.2.3`

### spec.replicas

`int32` · optional (explicit presence)

Number of Tempo replicas. 1 is the local-storage ceiling; with an
object-storage backend replicas share the backend and scale ingest
and query.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.storage

`KubernetesTempoStorage`

Where trace blocks live. Empty = local (a PersistentVolume).

### spec.storage.local

`KubernetesTempoLocalStorage`

Trace blocks on the pod's PersistentVolume. Single replica only.

### spec.storage.s3

`KubernetesTempoS3Storage`

Amazon S3 or any S3-compatible store. For in-cluster storage
point `endpoint` at a KubernetesSeaweedFs S3 endpoint with
`force_path_style: true`.

### spec.storage.s3.bucket

`string` · required

Bucket for trace blocks (must exist; Tempo does not create it).

- rule: {"required":true}

### spec.storage.s3.endpoint

`string` · required

S3 API endpoint host[:port] (e.g. "s3.us-east-1.amazonaws.com", or
"objects-filer.storage.svc.cluster.local:8333" for an in-cluster
KubernetesSeaweedFs). Required — Tempo does not derive it from the
region.

- rule: {"required":true}

### spec.storage.s3.region

`string`

AWS region (e.g. "us-east-1"). Usually empty for S3-compatible
endpoints.

### spec.storage.s3.forcePathStyle

`bool`

Use path-style addressing (bucket in the path, not the hostname).
Required by most S3-compatible endpoints, including SeaweedFS.

### spec.storage.s3.insecure

`bool`

Allow plain-HTTP endpoints. Only for in-cluster stores; never over
a network you do not own.

### spec.storage.s3.credentials

`KubernetesTempoObjectStoreCredentials`

Static credentials. Empty = the pod's ambient identity (IRSA on
EKS — the recommended keyless posture). When declared, the modules
inject them as environment variables from a Secret — they never
land in the rendered Tempo config.

### spec.storage.s3.credentials.accessKeyIdSecret

`KubernetesTempoSecretKeyRef` · required

Existing Secret holding the access key ID.

- rule: {"required":true}

### spec.storage.s3.credentials.accessKeyIdSecret.name

`string` · required

Secret name.

- rule: {"required":true}

### spec.storage.s3.credentials.accessKeyIdSecret.key

`string` · required

Key within the Secret.

- rule: {"required":true}

### spec.storage.s3.credentials.secretAccessKeySecret

`KubernetesTempoSecretKeyRef` · required

Existing Secret holding the secret access key.

- rule: {"required":true}

### spec.storage.s3.credentials.secretAccessKeySecret.name

`string` · required

Secret name.

- rule: {"required":true}

### spec.storage.s3.credentials.secretAccessKeySecret.key

`string` · required

Key within the Secret.

- rule: {"required":true}

### spec.storage.gcs

`KubernetesTempoGcsStorage`

Google Cloud Storage. With no key declared the pod's ambient
identity is used (GKE workload identity — the keyless posture).

### spec.storage.gcs.bucket

`string` · required

Bucket for trace blocks.

- rule: {"required":true}

### spec.storage.gcs.serviceAccountKeySecret

`KubernetesTempoSecretKeyRef`

Existing Secret holding a service-account JSON key. Empty = the
pod's ambient identity (GKE workload identity — the recommended
keyless posture). When declared, the key is mounted and referenced
via GOOGLE_APPLICATION_CREDENTIALS — never inlined in config.

### spec.storage.gcs.serviceAccountKeySecret.name

`string` · required

Secret name.

- rule: {"required":true}

### spec.storage.gcs.serviceAccountKeySecret.key

`string` · required

Key within the Secret.

- rule: {"required":true}

### spec.storage.azure

`KubernetesTempoAzureStorage`

Azure Blob Storage.

### spec.storage.azure.accountName

`string` · required

Storage-account name.

- rule: {"required":true}

### spec.storage.azure.container

`string` · required

Blob container for trace blocks.

- rule: {"required":true}

### spec.storage.azure.accountKeySecret

`KubernetesTempoSecretKeyRef`

Existing Secret holding the storage-account key. Empty = federated
workload identity (AKS — the recommended keyless posture).

### spec.storage.azure.accountKeySecret.name

`string` · required

Secret name.

- rule: {"required":true}

### spec.storage.azure.accountKeySecret.key

`string` · required

Key within the Secret.

- rule: {"required":true}

### spec.diskSize

`string` · optional (explicit presence)

Size of the persistent volume PER replica (e.g. "10Gi"). With
local storage this holds ALL trace blocks — size it for the full
retention window; with an object-storage backend it holds only the
WAL. Ignored when `ephemeral` is true.

- default: `10Gi`
- rule: {"string":{"pattern":"^\\d+(\\.\\d+)?\\s?(Ki|Mi|Gi|Ti|Pi|Ei|K|M|G|T|P|E)$"}}

### spec.storageClass

`string | valueFrom`

Storage class for the volumes. Accepts a literal name or a
reference to a KubernetesStorageClass resource. Empty = the
cluster's default class. Ignored when `ephemeral` is true.

- references: KubernetesStorageClass (`status.outputs.storage_class_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesStorageClass, name: <that resource's name>, fieldPath: status.outputs.storage_class_name}} -- a bare string does not parse

### spec.ephemeral

`bool`

Run WITHOUT persistent storage (the chart's own emptyDir default):
all locally-held traces and the WAL vanish with each pod restart.
Throwaway dev/test clusters only.

### spec.retention

`string` · optional (explicit presence)

How long trace blocks are kept before compaction deletes them, in
minutes or hours (e.g. "168h" for a week — Tempo parses Go
durations, which have no day unit). The chart default is 24h —
raise it for anything beyond a dev loop.

- default: `24h`
- rule: {"string":{"pattern":"^[0-9]+(m|h)$"}}

### spec.jaegerReceiversEnabled

`bool` · optional (explicit presence)

Also accept spans over the four legacy Jaeger protocols (gRPC
14250, thrift-binary 6832, thrift-compact 6831, thrift-http
14268). Default false — OTLP (always on: gRPC 4317, HTTP 4318) is
the 2026 wire standard, and every closed port is one less ingest
surface; this deliberately narrows the chart's all-receivers
default. Enable for fleets still emitting Jaeger.

- default: `false`

### spec.multiTenancyEnabled

`bool`

Require an X-Scope-OrgID tenant header on every push and query
(Tempo multi-tenancy). Clients and the Grafana datasource must
then send the header.

### spec.metricsGenerator

`KubernetesTempoMetricsGenerator`

Derive metrics from the trace stream (service graphs, span
metrics) and remote-write them to a Prometheus. Off by default.

- rule: an enabled metrics generator needs remote_write_url — the metrics it derives have nowhere to go without a Prometheus to write to

### spec.metricsGenerator.enabled

`bool`

Enable the metrics generator.

### spec.metricsGenerator.remoteWriteUrl

`string | valueFrom`

Prometheus endpoint the generated metrics are remote-written to.
Accepts a literal URL or a reference to a
KubernetesKubePrometheusStack resource (its Prometheus endpoint) —
the one-line wiring into the cluster's metrics. KNOW THIS: the
target Prometheus must accept pushes
(`prometheus.enable_remote_write_receiver: true` on the stack);
when the URL carries no path the modules append the standard
`/api/v1/write`.

- references: KubernetesKubePrometheusStack (`status.outputs.prometheus_endpoint`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesKubePrometheusStack, name: <that resource's name>, fieldPath: status.outputs.prometheus_endpoint}} -- a bare string does not parse

### spec.metricsGenerator.processors

`[]enum`

Which processors run. Empty = both service_graphs and
span_metrics.

- rule: {"repeated":{"unique":true,"items":{"enum":{"definedOnly":true,"notIn":[0]}}}}

Allowed values (use exactly as shown):

- `kubernetes_tempo_metrics_generator_processor_unspecified` -- Unspecified. Never valid in a manifest.
- `service_graphs` -- Build service-to-service request/latency metrics from trace spans — what draws Grafana's service map.
- `span_metrics` -- Per-span duration/count metrics (RED metrics per service and operation).

### spec.resources

`ContainerResources`

CPU and memory for the Tempo container. Empty = no requests/limits
(the chart default). Memory scales with ingest volume and block
size.

### spec.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.resources.limits.cpu

`string`

### spec.resources.limits.memory

`string`

### spec.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.resources.requests.cpu

`string`

### spec.resources.requests.memory

`string`

### spec.tempoQueryEnabled

`bool`

Deploy the tempo-query sidecar (the Jaeger-UI-compatible query
frontend on port 16686). Default false — Grafana is the query
surface; enable only for tooling that speaks the Jaeger API.

### spec.serviceMonitorEnabled

`bool`

Render a ServiceMonitor for Prometheus discovery. Requires the
monitoring.coreos.com CRDs on the cluster (deploy
KubernetesKubePrometheusStack first).

### spec.usageReporting

`bool` · optional (explicit presence)

Send anonymous usage statistics about this install to Grafana
Labs. Default false — this component deliberately diverges from
Tempo's report-by-default so no data leaves the cluster without an
explicit opt-in.

- default: `false`

### spec.imageRegistry

`string`

Registry that replaces the registry part of every image this
component pulls (tempo, tempo-query) — the air-gap/private-mirror
path. Empty = each image's upstream registry (docker.io).

### spec.imagePullSecrets

`[]string`

Names of existing image-pull Secrets applied to the Tempo pods.
The Secrets must already exist in the namespace.

### spec.scheduling

`KubernetesTempoScheduling`

Scheduling for the Tempo pods.

### spec.scheduling.nodeSelector

`map<string, string>`

Node selector for the Tempo pods.

### spec.scheduling.tolerations

`[]WorkloadToleration`

Tolerations for the Tempo pods.

### spec.scheduling.tolerations[].key

`string`

Taint key to tolerate. Empty key with operator "Exists" tolerates every taint.

### spec.scheduling.tolerations[].operator

`string`

How key/value match: "Equal" (default — value must match too) or "Exists"
(key presence alone matches).

- rule: Toleration operator must be either "Equal" or "Exists"

### spec.scheduling.tolerations[].value

`string`

Taint value to match when operator is "Equal".

### spec.scheduling.tolerations[].effect

`string`

Which taint effect is tolerated: "NoSchedule", "PreferNoSchedule", or
"NoExecute". Empty tolerates all effects for the key.

- rule: Toleration effect must be one of "NoSchedule", "PreferNoSchedule", or "NoExecute"

### spec.scheduling.tolerations[].tolerationSeconds

`int64` · optional (explicit presence)

For "NoExecute" taints only: how many seconds already-running pods stay bound
after the taint appears. Unset means tolerate forever.

### spec.scheduling.priorityClassName

`string`

Priority class name for the Tempo pods.

### spec.helmValues

`string`

Escape hatch: additional chart values as a YAML document, merged
LAST over everything the typed fields render (Helm `-f` semantics,
identical on both engines). For the chart surface beyond the typed
fields (per-receiver endpoints, tenant overrides, search tuning,
memberlist) — never the substitute for them. Do not put secrets
here; object-storage credentials belong in the typed
secret-reference fields.

## Validation Rules

- `spec.replicas.require_object_storage`: more than one replica requires an object-storage backend (s3, gcs or azure) — replicas cannot share local trace storage
- `spec.ephemeral.excludes_storage`: ephemeral: true runs on emptyDir — a non-default disk_size or a storage_class must not be set with it

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesTempo, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | namespace Tempo runs in. |
| `status.outputs.release_name` | `string` | Helm release name (= metadata.name). The modules pin the chart's fullname to it, so the Service name derives from it. |
| `status.outputs.service` | `string` | name of the Tempo Service (`<name>`, all ports). |
| `status.outputs.http_endpoint` | `string` | in-cluster HTTP endpoint (port 3200) — the URL Grafana `tempo` datasources and TraceQL clients use, e.g. http://traces.observability.svc.cluster.local:3200 |
| `status.outputs.otlp_grpc_endpoint` | `string` | in-cluster OTLP gRPC trace-ingest endpoint (port 4317) — where applications and KubernetesOtelCollector otlp exporters send spans, e.g. traces.observability.svc.cluster.local:4317 |
| `status.outputs.otlp_http_endpoint` | `string` | in-cluster OTLP HTTP trace-ingest endpoint (port 4318), e.g. http://traces.observability.svc.cluster.local:4318 |
| `status.outputs.port_forward_command` | `string` | command to port-forward the Tempo API to a developer laptop, e.g. kubectl port-forward svc/traces -n observability 3200:3200 |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.storageClass` | KubernetesStorageClass | `status.outputs.storage_class_name` |
| `spec.metricsGenerator.remoteWriteUrl` | KubernetesKubePrometheusStack | `status.outputs.prometheus_endpoint` |

## See Also

- [Overview](../README.md)
