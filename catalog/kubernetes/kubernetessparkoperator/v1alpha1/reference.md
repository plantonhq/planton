# KubernetesSparkOperator

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**KubernetesSparkOperatorSpec** deploys the Apache Spark Kubernetes
Operator — the official ASF controller that runs Spark workloads
declared as SparkApplication (one batch/streaming job, run to
completion) and SparkCluster (a long-lived standalone cluster)
custom resources — from the official `spark-kubernetes-operator`
Helm chart (https://apache.github.io/spark-kubernetes-operator).

GRAIN: one operator per cluster is the normal posture (it watches
cluster-wide by default). SparkApplication objects are per-JOB,
run-to-completion declarations — submit them per pipeline run
(typically from an orchestrator such as KubernetesAirflow), or
declare them via KubernetesManifest; they are deliberately not a
catalog kind.

CRD LIFECYCLE: the chart ships its two spark.apache.org CRDs
(sparkapplications, sparkclusters) in its `crds/` directory — Helm
installs them once, NEVER upgrades them on chart upgrades, and
LEAVES them (and every Spark workload declaration) on uninstall.
Apply the new release's CRD files manually when a chart bump
changes them.

NO ADMISSION WEBHOOK: this operator validates in its reconcile
loop — there is no webhook, no certificate machinery, and no
cert-manager dependency. A bad SparkApplication surfaces on the
CR's status, not as an admission rejection.

WORKLOAD RBAC IS PART OF THE INSTALL: Spark driver pods create and
delete executor pods at runtime, so every namespace that runs Spark
needs a service account with real pod-management permissions. The
chart owns that surface (see `workload`): cluster-scoped by
default, or fenced to an explicit namespace list — in which case
the chart CREATES those namespaces and the operator watches only
them.

OPERATOR CONFIGURATION is properties-based (`spark.kubernetes.
operator.*` keys rendered into a ConfigMap and appended over the
chart's defaults) — see `operator_properties` for the typed
surface and the upstream config_properties catalog for keys.

NAMING: keep the resource name at 40 characters or fewer — the
modules pin the chart's fullname AND every RBAC name to the
resource name (the chart hardcodes cluster-scoped RBAC names, so a
second install would otherwise collide), the longest derived name
suffixes "-config-monitor-binding" (23 chars), and Kubernetes
names cap at 63. Both engines fail loudly over the budget.

The typed fields cover the chart's meaningful surface;
`helm_values` remains the escape hatch (merged last, Helm `-f`
semantics, identical on both engines) — sentinel health canaries,
the operator NetworkPolicy, extra RBAC shapes — a safety valve,
never the primary interface.

## Example

```yaml
# Full-surface hack manifest for the offline plan/preview proofs: every
# module-rendered arm of the operator expressed at once — the fenced
# workload posture (namespaces list + service account), leader-elected
# replicas (exercising the module-owned leaderElection property), operator
# properties, hot-reload dynamic config, JVM args, resources, the
# image-registry mirror, pull secrets, full scheduling, and an
# escape-hatch document (sentinel canary) — while the module's RBAC name
# re-pins stay assertable in the rendered plan.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesSparkOperator
metadata:
  name: spark-operator-full
spec:
  namespace:
    value: spark-operator-system
  createNamespace: true
  chartVersion: 1.8.0
  replicas: 2
  workload:
    namespaces:
      - data-pipelines
      - ml-jobs
    serviceAccount: spark
  operatorProperties:
    spark.kubernetes.operator.reconciler.intervalSeconds: "30"
  dynamicConfig:
    enabled: true
    properties:
      spark.kubernetes.operator.reconciler.intervalSeconds: "60"
  jvmArgs: "-Dfile.encoding=UTF8 -XX:+UseParallelGC -XX:MaxRAMPercentage=80"
  resources:
    requests:
      cpu: 500m
      memory: 1Gi
    limits:
      cpu: "1"
      memory: 2Gi
  imageRegistry: mirror.example.com
  imagePullSecrets:
    - mirror-pull
  scheduling:
    nodeSelector:
      workload: system
    tolerations:
      - key: dedicated
        operator: Equal
        value: data-platform
        effect: NoSchedule
    priorityClassName: system-cluster-critical
  helmValues: |
    workloadResources:
      sparkApplicationSentinel:
        create: false
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.createNamespace` | `bool` |  |  |  |
| `spec.chartVersion` | `string` |  | `1.8.0` |  |
| `spec.replicas` | `int32` |  | `1` |  |
| `spec.workload` | `KubernetesSparkOperatorWorkload` |  |  |  |
| `spec.workload.namespaces` | `[]string` |  |  |  |
| `spec.workload.serviceAccount` | `string` |  | `spark` |  |
| `spec.operatorProperties` | `map<string, string>` |  |  |  |
| `spec.dynamicConfig` | `KubernetesSparkOperatorDynamicConfig` |  |  |  |
| `spec.dynamicConfig.enabled` | `bool` |  |  |  |
| `spec.dynamicConfig.properties` | `map<string, string>` |  |  |  |
| `spec.jvmArgs` | `string` |  |  |  |
| `spec.resources` | `ContainerResources` |  |  |  |
| `spec.resources.limits` | `CpuMemory` |  |  |  |
| `spec.resources.limits.cpu` | `string` |  |  |  |
| `spec.resources.limits.memory` | `string` |  |  |  |
| `spec.resources.requests` | `CpuMemory` |  |  |  |
| `spec.resources.requests.cpu` | `string` |  |  |  |
| `spec.resources.requests.memory` | `string` |  |  |  |
| `spec.imageRegistry` | `string` |  |  |  |
| `spec.imagePullSecrets` | `[]string` |  |  |  |
| `spec.scheduling` | `KubernetesSparkOperatorScheduling` |  |  |  |
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

Namespace to install the operator into. Accepts a literal
namespace name or a reference to a KubernetesNamespace resource.

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

Helm chart version to install (e.g. "1.8.0" — chart 1.8.0 pairs
with operator 1.0.0). Versions must exist as SERVED charts in the
repository index
(https://apache.github.io/spark-kubernetes-operator). NOTE chart
upgrades never upgrade the spark.apache.org CRDs (`crds/`
directory posture) — apply new CRD files manually when a bump
changes them.

- default: `1.8.0`

### spec.replicas

`int32` · optional (explicit presence)

Number of operator replicas. More than 1 turns on the operator's
leader election (one active reconciler, warm standbys) — the
modules set the leader-election properties for you; without them
the chart REFUSES multi-replica installs by design.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.workload

`KubernetesSparkOperatorWorkload`

Where Spark workloads are allowed to run — the operator's watch
scope AND the workload RBAC surface, decided together (they are
one chart surface). Empty = cluster-wide: the operator watches
every namespace and the chart creates a ClusterRole for Spark
pods. With `namespaces` set, the chart CREATES those namespaces,
plants the workload service account and a namespace-scoped Role
in each, and the operator watches ONLY them — the fenced,
multi-tenant posture.

### spec.workload.namespaces

`[]string`

Namespaces Spark workloads run in. Empty = cluster-wide watch
with a workload ClusterRole. Non-empty = the chart CREATES each
listed namespace (know this before pointing at names you manage
elsewhere), plants the service account + a namespace-scoped Role
in each, and the operator watches ONLY these namespaces —
SparkApplications anywhere else are ignored without an error.
NOTE the chart marks workload resources `helm.sh/resource-policy:
keep`, so these namespaces and their RBAC SURVIVE uninstall by
upstream design (running jobs must not lose their identity
mid-flight).

### spec.workload.serviceAccount

`string` · optional (explicit presence)

Service account name Spark driver/executor pods run as in every
workload namespace. Empty = "spark" (the upstream default —
SparkApplications reference it by this name).

- default: `spark`

### spec.operatorProperties

`map<string, string>`

Operator configuration properties appended over the chart's
defaults (`spark-operator.properties` — keys like
"spark.kubernetes.operator.reconciler.intervalSeconds"). The
full key catalog ships with the operator's docs at the pinned
version. Values are strings, exactly as the properties file
takes them.

### spec.dynamicConfig

`KubernetesSparkOperatorDynamicConfig`

Hot property reloading: the operator re-reads selected properties
from a ConfigMap at runtime instead of restarting. The modules
render the ConfigMap and the RBAC the chart pairs with it.
Off by default (upstream default) — most installs prefer
restart-on-change semantics.

### spec.dynamicConfig.enabled

`bool`

Enable hot property loading from a ConfigMap the modules render.

### spec.dynamicConfig.properties

`map<string, string>`

The hot-reloadable properties (same key space as
operator_properties). Changes to this map apply WITHOUT an
operator restart.

### spec.jvmArgs

`string`

JVM arguments for the operator container. Empty = the chart's
tuned default (parallel GC, 80% RAM percentage, crash on OOM).
Override only with a reason — the default is the upstream-tested
posture.

### spec.resources

`ContainerResources`

CPU and memory for the operator container. Empty = the chart
defaults (1 CPU / 2Gi — a JVM; the chart requests what it
limits). Lower this for lab clusters consciously: an OOM-killed
operator strands every reconciling Spark job.

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

### spec.imageRegistry

`string`

Registry that replaces the registry part of the operator image
(`apache/spark-kubernetes-operator`, Docker Hub implied) — the
air-gap/private-mirror path. This does NOT rewrite the images
Spark workloads run — those ride each SparkApplication's own
image field.

### spec.imagePullSecrets

`[]string`

Names of existing image-pull Secrets applied to the operator
pods. The Secrets must already exist in the namespace.

### spec.scheduling

`KubernetesSparkOperatorScheduling`

Scheduling for the operator pods.

### spec.scheduling.nodeSelector

`map<string, string>`

Node selector for the operator pods.

### spec.scheduling.tolerations

`[]WorkloadToleration`

Tolerations for the operator pods.

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

Priority class name for the operator pods.

### spec.helmValues

`string`

Escape hatch: additional chart values as a YAML document, merged
LAST over everything the typed fields render (Helm `-f`
semantics, identical on both engines). For the chart surface
beyond the typed fields (sentinel canaries, NetworkPolicy,
probe tuning, RBAC name overrides) — never the substitute for
them.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesSparkOperator, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | namespace the operator runs in. |
| `status.outputs.release_name` | `string` | Helm release name (= metadata.name). The modules pin the chart's fullname to it, so child names derive from it. |
| `status.outputs.workload_service_account` | `string` | service account name Spark driver/executor pods run as in every workload namespace — SparkApplication declarations reference it. |
| `status.outputs.watched_namespaces` | `[]string` | namespaces the operator watches for Spark workloads. Empty means cluster-wide — SparkApplications reconcile anywhere. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |

## See Also

- [Overview](../README.md)
