# KubernetesNeo4j

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

**KubernetesNeo4jSpec** deploys a Neo4j graph database — the
standard engine for knowledge graphs, GraphRAG and agent-memory
architectures — from the official `neo4j` Helm chart
(https://helm.neo4j.com/neo4j).

GRAIN: one release = one Neo4j server (the chart's own grain — a
StatefulSet of exactly one pod). Community edition (the default)
is single-instance by license; Enterprise clustering is built by
installing MULTIPLE KubernetesNeo4j resources that share the same
`cluster_name` — each member is its own first-class resource, not
a replicas knob.

CREDENTIALS ARE MINTED, NEVER REQUIRED: leave `auth` empty and the
modules generate the `neo4j` admin user's password, materialize it as
the `<name>-auth` Secret (the chart's `NEO4J_AUTH: neo4j/<password>`
contract plus a bare `password` key), and point the chart at it — the
password never lands in rendered Helm values, and `auth_secret_name` /
`password_secret` tell workloads where to read it. Declare
`auth.password` (a managed-secret reference) to bring your own, or
`auth.existing_secret` to reference a Secret you own that already
carries the chart's contract.

EXPOSURE: the chart's default LoadBalancer service is deliberately
overridden to ClusterIP — exposure composes from first-class kinds
(KubernetesIngress, Gateway API kinds) over the exported service
handle, or set `service.type` for a quick LoadBalancer.

The typed fields below cover the chart's meaningful configuration
surface; `helm_values` remains as the escape hatch for chart
values beyond them (merged last, Helm `-f` semantics, identical on
both engines) — a safety valve, never the primary interface.

## Example

```yaml
# Full-surface offline-proof manifest: exercises the community edition, the
# password auth arm (materialized as the neo4j-hack-auth Secret), a
# dynamically-provisioned data volume with an explicit StorageClass, memory
# tuning, extra neo4j.conf and apoc.conf entries, additional JVM arguments,
# the ClusterIP service default with annotations, scheduling, and the
# ServiceMonitor toggle (off) — so the offline tofu plan and pulumi preview
# proofs cover the full typed surface. Resources honour the chart's floor
# (500m CPU / 2Gi memory — the chart rejects installs below it).
# Placeholder values; never applied to a real cluster.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesNeo4j
metadata:
  name: neo4j-hack
spec:
  namespace:
    value: neo4j-hack
  createNamespace: true
  chartVersion: 2026.6.0
  edition: community
  auth:
    password: hack-placeholder-password
  resources:
    requests:
      cpu: 500m
      memory: 2Gi
    limits:
      cpu: "2"
      memory: 4Gi
  dataVolume:
    size: 20Gi
    storageClass:
      value: fast-ssd
  memory:
    heapInitial: 1G
    heapMax: 1G
    pageCache: 512M
  config:
    db.transaction.timeout: 30s
    server.bolt.thread_pool_max_size: "400"
  apocConfig:
    apoc.trigger.enabled: "true"
    apoc.import.file.enabled: "true"
  additionalJvmArguments:
    - "-XX:+HeapDumpOnOutOfMemoryError"
    - "-XX:HeapDumpPath=/logs/neo4j.hprof"
  useDefaultJvmArguments: true
  service:
    type: ClusterIP
    annotations:
      example.org/purpose: knowledge-graph
  scheduling:
    nodeSelector:
      kubernetes.io/os: linux
    tolerations:
      - key: dedicated
        operator: Equal
        value: graph
        effect: NoSchedule
    podAntiAffinity: true
    priorityClassName: system-cluster-critical
  serviceMonitorEnabled: false
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.createNamespace` | `bool` |  |  |  |
| `spec.chartVersion` | `string` |  | `2026.6.0` |  |
| `spec.edition` | `string` |  | `community` |  |
| `spec.acceptLicenseAgreement` | `bool` |  |  |  |
| `spec.auth` | `KubernetesNeo4jAuth` |  |  |  |
| `spec.auth.password` | `string` (sensitive) |  |  |  |
| `spec.auth.existingSecret` | `string` |  |  |  |
| `spec.clusterName` | `string` |  |  |  |
| `spec.resources` | `ContainerResources` |  |  |  |
| `spec.resources.limits` | `CpuMemory` |  |  |  |
| `spec.resources.limits.cpu` | `string` |  |  |  |
| `spec.resources.limits.memory` | `string` |  |  |  |
| `spec.resources.requests` | `CpuMemory` |  |  |  |
| `spec.resources.requests.cpu` | `string` |  |  |  |
| `spec.resources.requests.memory` | `string` |  |  |  |
| `spec.dataVolume` | `KubernetesNeo4jDataVolume` |  |  |  |
| `spec.dataVolume.size` | `string` |  | `10Gi` |  |
| `spec.dataVolume.storageClass` | `string \| valueFrom` |  |  | KubernetesStorageClass (`status.outputs.storage_class_name`) |
| `spec.memory` | `KubernetesNeo4jMemory` |  |  |  |
| `spec.memory.heapInitial` | `string` |  |  |  |
| `spec.memory.heapMax` | `string` |  |  |  |
| `spec.memory.pageCache` | `string` |  |  |  |
| `spec.config` | `map<string, string>` |  |  |  |
| `spec.apocConfig` | `map<string, string>` |  |  |  |
| `spec.additionalJvmArguments` | `[]string` |  |  |  |
| `spec.useDefaultJvmArguments` | `bool` |  | `true` |  |
| `spec.service` | `KubernetesNeo4jService` |  |  |  |
| `spec.service.type` | `string` |  | `ClusterIP` |  |
| `spec.service.annotations` | `map<string, string>` |  |  |  |
| `spec.ssl` | `KubernetesNeo4jSsl` |  |  |  |
| `spec.ssl.bolt` | `KubernetesNeo4jSslScope` |  |  |  |
| `spec.ssl.bolt.secret` | `string \| valueFrom` |  |  | KubernetesCertificate (`status.outputs.secret_name`) |
| `spec.ssl.https` | `KubernetesNeo4jSslScope` |  |  |  |
| `spec.ssl.https.secret` | `string \| valueFrom` |  |  | KubernetesCertificate (`status.outputs.secret_name`) |
| `spec.scheduling` | `KubernetesNeo4jScheduling` |  |  |  |
| `spec.scheduling.nodeSelector` | `map<string, string>` |  |  |  |
| `spec.scheduling.tolerations` | `[]WorkloadToleration` |  |  |  |
| `spec.scheduling.tolerations[].key` | `string` |  |  |  |
| `spec.scheduling.tolerations[].operator` | `string` |  |  |  |
| `spec.scheduling.tolerations[].value` | `string` |  |  |  |
| `spec.scheduling.tolerations[].effect` | `string` |  |  |  |
| `spec.scheduling.tolerations[].tolerationSeconds` | `int64` |  |  |  |
| `spec.scheduling.podAntiAffinity` | `bool` |  | `true` |  |
| `spec.scheduling.priorityClassName` | `string` |  |  |  |
| `spec.serviceMonitorEnabled` | `bool` |  |  |  |
| `spec.image` | `KubernetesNeo4jImage` |  |  |  |
| `spec.image.registry` | `string` |  |  |  |
| `spec.image.repository` | `string` |  |  |  |
| `spec.image.tag` | `string` |  |  |  |
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
governance labels) before installing and deleted with the
resource. When false, the namespace must already exist.

### spec.chartVersion

`string` · optional (explicit presence)

Helm chart version to install (e.g. "2026.6.0" — chart versions
track Neo4j calendar releases). Versions must exist as SERVED
charts in the repository index (https://helm.neo4j.com/neo4j).

- default: `2026.6.0`

### spec.edition

`string` · optional (explicit presence)

Neo4j edition. Community (default, GPLv3, single-instance) or
Enterprise (commercial license, enables clustering and advanced
features; requires `accept_license_agreement` and a valid Neo4j
license).

- default: `community`
- rule: edition must be community or enterprise

### spec.acceptLicenseAgreement

`bool`

Accept the Neo4j license agreement — REQUIRED for the enterprise
edition (renders the chart's acceptLicenseAgreement: "yes"). Has
no effect on community.

### spec.auth

`KubernetesNeo4jAuth`

Admin (`neo4j` user) credentials. Empty = the module generates a
password and materializes it as the `<name>-auth` Secret exactly as
it would a declared one — the recommended shape, since nobody has to
mint or escrow a value the module can. Declare an arm only to bring
your own.

### spec.auth.password

`string` · sensitive

The admin password, declared here (as a managed-secret reference)
and materialized by the modules as the `<name>-auth` Secret (key
NEO4J_AUTH, the chart's contract, plus a bare `password` key).
Never appears in rendered Helm values.

### spec.auth.existingSecret

`string`

Name of an existing Secret already carrying the chart's
NEO4J_AUTH key ("neo4j/<password>"). The Secret must exist
BEFORE the install — the chart reads it at template time.

### spec.clusterName

`string`

Server name for clustering: Enterprise members that share this
name form one cluster. Empty = standalone (and always standalone
on community).

### spec.resources

`ContainerResources`

Container resources. Chart minimum: 500m CPU / 2Gi memory —
the chart REJECTS installs below its floor.

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

### spec.dataVolume

`KubernetesNeo4jDataVolume`

Data volume. Empty = a 10Gi PVC on the cluster's default
StorageClass.

### spec.dataVolume.size

`string` · optional (explicit presence)

Volume size, e.g. "10Gi".

- default: `10Gi`

### spec.dataVolume.storageClass

`string | valueFrom`

StorageClass for the data PVC. Accepts a literal class name or a
reference to a KubernetesStorageClass resource. Empty = the
cluster's default class.

- references: KubernetesStorageClass (`status.outputs.storage_class_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesStorageClass, name: <that resource's name>, fieldPath: status.outputs.storage_class_name}} -- a bare string does not parse

### spec.memory

`KubernetesNeo4jMemory`

Memory tuning rendered into neo4j.conf. Empty = Neo4j
auto-computes from the container memory (the chart default —
usually right; set explicitly for shared or memory-tight nodes).

### spec.memory.heapInitial

`string`

Initial JVM heap, e.g. "1G" (server.memory.heap.initial_size).

### spec.memory.heapMax

`string`

Max JVM heap, e.g. "2G" (server.memory.heap.max_size). Keep
initial = max for production.

### spec.memory.pageCache

`string`

Page cache for graph data, e.g. "1G"
(server.memory.pagecache.size). Rule of thumb: what remains
after heap and OS overhead.

### spec.config

`map<string, string>`

Extra neo4j.conf entries, exactly as neo4j.conf expects them
(e.g. "server.default_listen_address": "0.0.0.0"). Memory keys
belong in `memory`; auth/TLS keys are chart-owned.

### spec.apocConfig

`map<string, string>`

APOC plugin configuration entries (apoc.conf).

### spec.additionalJvmArguments

`[]string`

Additional JVM arguments (appended to the chart's defaults
unless use_default_jvm_arguments is false).

### spec.useDefaultJvmArguments

`bool` · optional (explicit presence)

Keep the chart's default JVM arguments. Chart default: true.

- default: `true`

### spec.service

`KubernetesNeo4jService`

The chart's EXPOSURE Service (`<name>-lb-neo4j`; bolt 7687,
http 7474, https 7473). Empty = ClusterIP — NOTE this is a
deliberate override of the chart's LoadBalancer default; exposure
composes from first-class kinds instead. In-cluster clients use
the always-created default Service (= the resource name — the
endpoints in the stack outputs), so this block matters only when
exposing the server directly.

### spec.service.type

`string` · optional (explicit presence)

Service type: ClusterIP (the component default — NOT the chart's
LoadBalancer default), NodePort, or LoadBalancer.

- default: `ClusterIP`
- rule: service type must be ClusterIP, NodePort, or LoadBalancer

### spec.service.annotations

`map<string, string>`

Service annotations (cloud LB recipes ride here when type is
LoadBalancer).

### spec.ssl

`KubernetesNeo4jSsl`

Server-side TLS from existing certificate Secrets, per scope.
Empty = plaintext in-cluster (compose TLS at the exposure layer,
or declare bolt/https certificates here).

### spec.ssl.bolt

`KubernetesNeo4jSslScope`

TLS for the bolt (7687) listener.

### spec.ssl.bolt.secret

`string | valueFrom`

Existing TLS Secret (private.key + public.crt). Accepts a
literal name or a KubernetesCertificate reference — the
cert-manager seam (cert-manager Secrets carry tls.key/tls.crt;
see the component docs for the key-name bridge).

- references: KubernetesCertificate (`status.outputs.secret_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesCertificate, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.ssl.https

`KubernetesNeo4jSslScope`

TLS for the https (7473) listener.

### spec.ssl.https.secret

`string | valueFrom`

Existing TLS Secret (private.key + public.crt). Accepts a
literal name or a KubernetesCertificate reference — the
cert-manager seam (cert-manager Secrets carry tls.key/tls.crt;
see the component docs for the key-name bridge).

- references: KubernetesCertificate (`status.outputs.secret_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesCertificate, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.scheduling

`KubernetesNeo4jScheduling`

Scheduling for the server pod.

### spec.scheduling.nodeSelector

`map<string, string>`

Node selector for the server pod.

### spec.scheduling.tolerations

`[]WorkloadToleration`

Tolerations for the server pod.

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

### spec.scheduling.podAntiAffinity

`bool` · optional (explicit presence)

Pod anti-affinity against other Neo4j pods. Chart default: true
(meaningful for enterprise cluster members; harmless
standalone).

- default: `true`

### spec.scheduling.priorityClassName

`string`

Priority class name for the server pod.

### spec.serviceMonitorEnabled

`bool`

Create a ServiceMonitor for Prometheus scraping (requires the
Prometheus Operator CRDs). Chart default: false.

### spec.image

`KubernetesNeo4jImage`

Override the Neo4j image (air-gap path). Empty = the chart's
official image for `chart_version` and `edition`.

### spec.image.registry

`string`

Image registry, e.g. "my.private.registry.com". Empty = Docker
Hub.

### spec.image.repository

`string`

Image repository. Empty = "neo4j".

### spec.image.tag

`string`

Image tag. Empty = the chart's default for the pinned version
and edition.

### spec.helmValues

`string`

Escape hatch: additional chart values as a YAML document, merged
LAST over everything the typed fields render (Helm `-f`
semantics, identical on both engines). For the chart surface
beyond the typed fields (log4j XML, additional volumes/mounts,
LDAP secrets, the operations sidecar, probes, PDB, per-service
splits, ...) — never the substitute for them. Do not put secrets
here.

## Validation Rules

- `spec.enterprise_requires_license_acceptance`: the enterprise edition requires accept_license_agreement: true (and a valid Neo4j license)
- `spec.cluster_requires_enterprise`: cluster_name requires the enterprise edition — community Neo4j is single-instance by license

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesNeo4j, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | namespace the server runs in. |
| `status.outputs.release_name` | `string` | Helm release name (= metadata.name). |
| `status.outputs.service_name` | `string` | name of the main Neo4j Service (bolt/http/https ports). |
| `status.outputs.bolt_endpoint` | `string` | in-cluster bolt endpoint drivers connect to, e.g. neo4j://main.graph.svc.cluster.local:7687 |
| `status.outputs.http_endpoint` | `string` | in-cluster HTTP API endpoint, e.g. http://main.graph.svc.cluster.local:7474 |
| `status.outputs.auth_secret_name` | `string` | name of the Secret holding the admin credentials in the chart's contract (key NEO4J_AUTH, value "neo4j/<password>"): the module-materialized `<name>-auth` — whether the password was declared or generated by the module — or the referenced existing Secret. Always set. |
| `status.outputs.port_forward_command` | `string` | command to port-forward bolt to a developer laptop, e.g. kubectl port-forward svc/main -n graph 7687:7687 |
| `status.outputs.password_secret` | `KubernetesSecretKey` | the Secret key holding the `neo4j` admin user's BARE password (the module-materialized `<name>-auth` Secret's `password` key), for workloads that take a password rather than the chart's "neo4j/<password>" pair. Unset when auth.existing_secret is declared: that Secret's layout is the owner's, and the module never assumes it carries a bare-password key. |
| `status.outputs.password_secret.name` | `string` | The name of the Kubernetes Secret. |
| `status.outputs.password_secret.key` | `string` | The key within the Kubernetes Secret. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.dataVolume.storageClass` | KubernetesStorageClass | `status.outputs.storage_class_name` |
| `spec.ssl.bolt.secret` | KubernetesCertificate | `status.outputs.secret_name` |
| `spec.ssl.https.secret` | KubernetesCertificate | `status.outputs.secret_name` |

## See Also

- [Overview](../README.md)
