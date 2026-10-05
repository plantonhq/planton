# KubernetesPerconaMongoOperator

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**KubernetesPerconaMongoOperatorSpec** installs the Percona Operator
for MongoDB from the official `psmdb-operator` Helm chart
(https://percona.github.io/percona-helm-charts). The operator
reconciles `PerconaServerMongoDB` custom resources into highly
available MongoDB deployments: replica sets with automated failover,
sharded clusters (mongos + config servers), scheduled logical,
physical, and incremental backups with point-in-time recovery via
Percona Backup for MongoDB, TLS, and user management.

This component installs and configures the ENGINE. The databases
themselves are declared with KubernetesMongodb resources — one per
MongoDB cluster — which the operator reconciles.

WATCH SCOPE: by default the operator watches ITS OWN namespace only
(the upstream default — databases live beside their operator). Set
`watch.cluster_wide` for one operator managing databases in every
namespace, or `watch.namespaces` for a fenced set.

CRD LIFECYCLE: the chart ships the PerconaServerMongoDB CRDs in its
Helm-native `crds/` directory — installed on first install, never
upgraded or deleted by Helm. Uninstalling the release therefore NEVER
cascade-deletes the database clusters (the upstream safety posture).
The same posture means a `chart_version` upgrade runs new operator
code against the EXISTING CRDs — apply the new release's CRDs
yourself when an upgrade's release notes call for it.

The typed fields below cover the chart's meaningful configuration
surface; `helm_values` remains as the escape hatch for chart values
beyond them (merged last, Helm `-f` semantics, identical on both
engines) — a safety valve, never the primary interface.

## Example

```yaml
# Full-surface test manifest: exercises every typed arm of the spec so the
# offline plan proofs cover what the live lanes may not. Not a realistic
# production shape — see presets for those.
#
# watch.namespaces is set (and watch.clusterWide left false): the spec CEL
# rules make the two arms mutually exclusive, and the namespace fence
# exercises the comma-joined watchNamespace rendering — the richer of the
# two mappings.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesPerconaMongoOperator
metadata:
  name: hack-psmdb-operator
spec:
  namespace:
    value: hack-psmdb-system
  createNamespace: true
  chartVersion: "1.22.0"
  replicas: 2
  watch:
    namespaces:
      - team-a
      - team-b
  maxConcurrentReconciles: 5
  log:
    structured: true
    level: DEBUG
  disableTelemetry: true
  resources:
    requests:
      cpu: 100m
      memory: 128Mi
    limits:
      cpu: "1"
      memory: 512Mi
  nodeSelector:
    kubernetes.io/os: linux
  tolerations:
    - key: node-role.kubernetes.io/control-plane
      operator: Exists
      effect: NoSchedule
    - key: dedicated
      operator: Equal
      value: databases
      effect: NoExecute
      tolerationSeconds: 300
  imagePullSecrets:
    - hack-registry-credentials
  image:
    repository: mirror.example.com/percona/percona-server-mongodb-operator
    tag: 1.22.0
  helmValues: |
    podAnnotations:
      prometheus.io/scrape: "true"
      prometheus.io/port: "8080"
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.createNamespace` | `bool` |  |  |  |
| `spec.chartVersion` | `string` |  | `1.22.0` |  |
| `spec.replicas` | `int32` |  | `1` |  |
| `spec.watch` | `KubernetesPerconaMongoOperatorWatch` |  |  |  |
| `spec.watch.clusterWide` | `bool` |  |  |  |
| `spec.watch.namespaces` | `[]string` |  |  |  |
| `spec.maxConcurrentReconciles` | `int32` |  | `1` |  |
| `spec.log` | `KubernetesPerconaMongoOperatorLog` |  |  |  |
| `spec.log.structured` | `bool` |  |  |  |
| `spec.log.level` | `string` |  | `INFO` |  |
| `spec.disableTelemetry` | `bool` |  |  |  |
| `spec.resources` | `ContainerResources` |  |  |  |
| `spec.resources.limits` | `CpuMemory` |  |  |  |
| `spec.resources.limits.cpu` | `string` |  |  |  |
| `spec.resources.limits.memory` | `string` |  |  |  |
| `spec.resources.requests` | `CpuMemory` |  |  |  |
| `spec.resources.requests.cpu` | `string` |  |  |  |
| `spec.resources.requests.memory` | `string` |  |  |  |
| `spec.nodeSelector` | `map<string, string>` |  |  |  |
| `spec.tolerations` | `[]WorkloadToleration` |  |  |  |
| `spec.tolerations[].key` | `string` |  |  |  |
| `spec.tolerations[].operator` | `string` |  |  |  |
| `spec.tolerations[].value` | `string` |  |  |  |
| `spec.tolerations[].effect` | `string` |  |  |  |
| `spec.tolerations[].tolerationSeconds` | `int64` |  |  |  |
| `spec.imagePullSecrets` | `[]string` |  |  |  |
| `spec.image` | `KubernetesPerconaMongoOperatorImage` |  |  |  |
| `spec.image.repository` | `string` |  |  |  |
| `spec.image.tag` | `string` |  |  |  |
| `spec.helmValues` | `string` |  |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

Namespace to install the operator into. Accepts a literal namespace
name or a reference to a KubernetesNamespace resource.

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

Helm chart version to install (e.g. "1.22.0" — chart and operator
versions move together for this chart; the chart pin governs). Pin
deliberately; upgrades re-run the release with the new chart.

- default: `1.22.0`

### spec.replicas

`int32` · optional (explicit presence)

Operator replica count. Chart default: 1. Extra replicas are
leader-elected warm standbys for the OPERATOR itself — they add no
reconciliation throughput.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.watch

`KubernetesPerconaMongoOperatorWatch`

What the operator watches. Omitted = its own namespace only (the
upstream default).

- rule: cluster_wide and namespaces are mutually exclusive — a cluster-wide operator already watches everything

### spec.watch.clusterWide

`bool`

Watch every namespace in the cluster (cluster-wide RBAC). Chart
default: false.

### spec.watch.namespaces

`[]string`

Watch exactly these namespaces (rendered as the chart's
comma-separated watchNamespace). Empty with cluster_wide false =
the operator's own namespace only.

### spec.maxConcurrentReconciles

`int32` · optional (explicit presence)

Maximum number of PerconaServerMongoDB resources reconciled
concurrently. Chart default: 1. Raise on control planes managing
many databases.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.log

`KubernetesPerconaMongoOperatorLog`

Operator logging.

### spec.log.structured

`bool`

Emit structured (JSON) logs instead of the console encoder. Chart
default: false.

### spec.log.level

`string` · optional (explicit presence)

Log level: DEBUG, INFO (the chart default), or ERROR.

- default: `INFO`
- rule: log level must be DEBUG, INFO, or ERROR

### spec.disableTelemetry

`bool`

Disable Percona's anonymous telemetry (version and feature usage
pings to check.percona.com). Chart default: false (telemetry on).

### spec.resources

`ContainerResources`

Operator container resources. Empty = no requests/limits (the chart
ships none by default).

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

### spec.nodeSelector

`map<string, string>`

Node selector for the operator pod.

### spec.tolerations

`[]WorkloadToleration`

Tolerations for the operator pod.

### spec.tolerations[].key

`string`

Taint key to tolerate. Empty key with operator "Exists" tolerates every taint.

### spec.tolerations[].operator

`string`

How key/value match: "Equal" (default — value must match too) or "Exists"
(key presence alone matches).

- rule: Toleration operator must be either "Equal" or "Exists"

### spec.tolerations[].value

`string`

Taint value to match when operator is "Equal".

### spec.tolerations[].effect

`string`

Which taint effect is tolerated: "NoSchedule", "PreferNoSchedule", or
"NoExecute". Empty tolerates all effects for the key.

- rule: Toleration effect must be one of "NoSchedule", "PreferNoSchedule", or "NoExecute"

### spec.tolerations[].tolerationSeconds

`int64` · optional (explicit presence)

For "NoExecute" taints only: how many seconds already-running pods stay bound
after the taint appears. Unset means tolerate forever.

### spec.imagePullSecrets

`[]string`

Names of image-pull secrets (in the installation namespace) for
pulling the operator image from a private mirror.

### spec.image

`KubernetesPerconaMongoOperatorImage`

Override the operator image (registry mirrors, air-gapped
clusters). Empty = the chart default
(percona/percona-server-mongodb-operator at the chart version).

### spec.image.repository

`string`

Image repository (e.g.
"my-mirror.example.com/percona-server-mongodb-operator").

### spec.image.tag

`string`

Image tag. Empty = the chart version.

### spec.helmValues

`string`

Escape hatch: additional chart values as a YAML document, merged
LAST over everything the typed fields render (Helm `-f` semantics,
identical on both engines). For the chart surface beyond the typed
fields (security contexts, RBAC toggles, resync period, pod
annotations, ...) — never the substitute for them. Do not put
secrets here.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesPerconaMongoOperator, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | Namespace the operator runs in. |
| `status.outputs.release_name` | `string` | Helm release name of the operator (`metadata.name`). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |

## See Also

- [Overview](../README.md)
