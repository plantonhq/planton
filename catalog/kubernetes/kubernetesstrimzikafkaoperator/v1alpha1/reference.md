# KubernetesStrimziKafkaOperator

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**KubernetesStrimziKafkaOperatorSpec** installs the Strimzi cluster
operator — the CNCF project that runs Apache Kafka on Kubernetes —
from the official `strimzi-kafka-operator` Helm chart
(https://strimzi.io/charts/). The operator reconciles `Kafka` custom
resources (declared with KubernetesKafka) into KRaft-mode Kafka
clusters, and its per-cluster entity operators reconcile
`KafkaTopic` / `KafkaUser` resources (KubernetesKafkaTopic /
KubernetesKafkaUser) into real topics and authenticated users.

This component installs and configures the ENGINE. Kafka clusters
themselves are declared with KubernetesKafka resources — one per
cluster — which this operator reconciles.

WATCH SCOPE: by default the operator watches ITS OWN namespace only
(the chart default — Kafka clusters live beside their operator). Set
`watch.any_namespace` for one operator managing Kafka clusters in
every namespace, or `watch.namespaces` for a fenced set.

CRD LIFECYCLE: the chart ships the Strimzi CRDs (Kafka,
KafkaNodePool, KafkaTopic, KafkaUser, ...) in its Helm-native `crds/`
directory — installed on first install, never upgraded or deleted by
Helm. Uninstalling the release therefore NEVER cascade-deletes Kafka
clusters (the upstream safety posture). The same posture means a
`chart_version` upgrade runs new operator code against the EXISTING
CRDs — apply the new release's CRDs yourself when an upgrade's
release notes call for it (Strimzi minor upgrades regularly add CRD
fields).

MULTIPLE INSTALLS: a second operator in the same cluster (watching a
disjoint namespace set) must set `create_global_resources` false in
`helm_values` terms — the chart's fixed-name ClusterRoles are owned
by the first release and a second install conflicts on them.

The typed fields below cover the chart's meaningful configuration
surface; `helm_values` remains as the escape hatch for chart values
beyond them (merged last, Helm `-f` semantics, identical on both
engines) — a safety valve, never the primary interface.

## Example

```yaml
# Full-surface development manifest: exercises every typed field so the
# offline plan proofs cover arms the live lanes exclude. Values are
# deliberately non-default where a default exists, so rendered-values
# inspection shows every mapping.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesStrimziKafkaOperator
metadata:
  name: strimzi-operator-hack
spec:
  namespace:
    value: strimzi-hack
  createNamespace: true
  chartVersion: 1.1.0
  replicas: 2
  watch:
    namespaces:
      - kafka-team-a
      - kafka-team-b
  fullReconciliationIntervalMs: 180000
  operationTimeoutMs: 420000
  logLevel: DEBUG
  # featureGates deliberately empty: an unknown gate name fails operator
  # startup, and the pinned release needs no gate flips.
  kubernetesServiceDnsDomain: cluster.local
  leaderElectionEnabled: true
  generateNetworkPolicy: false
  generatePodDisruptionBudget: false
  createGlobalResources: true
  resources:
    requests:
      cpu: 100m
      memory: 256Mi
    limits:
      cpu: 500m
      memory: 512Mi
  nodeSelector:
    kubernetes.io/os: linux
  tolerations:
    - key: dedicated
      operator: Equal
      value: kafka
      effect: NoSchedule
  imagePullSecrets:
    - mirror-pull-secret
  image:
    registry: mirror.example.com
    repository: strimzi
    tag: 1.1.0
  helmValues: |
    dashboards:
      enabled: false
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.createNamespace` | `bool` |  |  |  |
| `spec.chartVersion` | `string` |  | `1.1.0` |  |
| `spec.replicas` | `int32` |  | `1` |  |
| `spec.watch` | `KubernetesStrimziKafkaOperatorWatch` |  |  |  |
| `spec.watch.anyNamespace` | `bool` |  |  |  |
| `spec.watch.namespaces` | `[]string` |  |  |  |
| `spec.fullReconciliationIntervalMs` | `int32` |  | `120000` |  |
| `spec.operationTimeoutMs` | `int32` |  | `300000` |  |
| `spec.logLevel` | `string` |  | `INFO` |  |
| `spec.featureGates` | `string` |  |  |  |
| `spec.kubernetesServiceDnsDomain` | `string` |  | `cluster.local` |  |
| `spec.leaderElectionEnabled` | `bool` |  | `true` |  |
| `spec.generateNetworkPolicy` | `bool` |  | `true` |  |
| `spec.generatePodDisruptionBudget` | `bool` |  | `true` |  |
| `spec.createGlobalResources` | `bool` |  | `true` |  |
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
| `spec.image` | `KubernetesStrimziKafkaOperatorImage` |  |  |  |
| `spec.image.registry` | `string` |  |  |  |
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

Helm chart version to install (e.g. "1.1.0" — chart and operator
versions move together for this chart; the chart pin governs).
Versions must exist as SERVED charts in the repository index
(https://strimzi.io/charts/) — the Chart.yaml inside the Strimzi
source tree carries a build-time placeholder and never reflects
the served version. Remember the `crds/`-directory posture: a
version upgrade re-runs the release with the new chart but leaves
the CRDs at their installed versions.

- default: `1.1.0`

### spec.replicas

`int32` · optional (explicit presence)

Operator replica count. Chart default: 1. Extra replicas are
leader-elected warm standbys for the OPERATOR itself — they add no
reconciliation throughput.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.watch

`KubernetesStrimziKafkaOperatorWatch`

What the operator watches. Omitted = its own namespace only (the
chart default).

- rule: any_namespace and namespaces are mutually exclusive — a cluster-wide operator already watches everything

### spec.watch.anyNamespace

`bool`

Watch every namespace in the cluster (rendered as the chart's
watchAnyNamespace, cluster-wide RBAC). Chart default: false.

### spec.watch.namespaces

`[]string`

Watch exactly these namespaces (rendered as the chart's
watchNamespaces list — the installation namespace is always
watched in addition). Empty with any_namespace false = the
operator's own namespace only.

Every listed namespace MUST already exist: the chart templates
RoleBindings into each one, and the install fails with
"namespaces ... not found" otherwise (verified live — the chart
does not create watched namespaces). Create them first, e.g. as
KubernetesNamespace resources the install depends on.

### spec.fullReconciliationIntervalMs

`int32` · optional (explicit presence)

Interval between full reconciliations of every watched resource,
in milliseconds. Chart default: 120000 (2 minutes). The periodic
pass repairs drift that event-driven reconciliation missed.

- default: `120000`
- rule: {"int32":{"gte":1000}}

### spec.operationTimeoutMs

`int32` · optional (explicit presence)

Timeout for internal operations (rolling a pod, waiting for a
readiness gate) in milliseconds. Chart default: 300000
(5 minutes). Raise on slow storage where broker restarts
legitimately exceed it — an expired timeout fails the whole
reconciliation.

- default: `300000`
- rule: {"int32":{"gte":1000}}

### spec.logLevel

`string` · optional (explicit presence)

Operator log level: ERROR, WARN, INFO (the chart default), DEBUG,
or TRACE.

- default: `INFO`
- rule: log level must be ERROR, WARN, INFO, DEBUG, or TRACE

### spec.featureGates

`string`

Strimzi feature gates, in the operator's own syntax: a
comma-separated list of gate names each prefixed with `+` (enable)
or `-` (disable), e.g. "+DummyGate,-OtherGate". Empty = the
release defaults. Consult the pinned Strimzi release's
documentation for the gates it recognizes — an unknown gate fails
operator startup.

### spec.kubernetesServiceDnsDomain

`string` · optional (explicit presence)

Kubernetes cluster DNS domain the operator bakes into generated
addresses (advertised listeners, certificates). Chart default:
"cluster.local". Set only on clusters with a non-default DNS
domain — a mismatch produces TLS certificates whose SANs do not
match the service DNS names brokers advertise.

- default: `cluster.local`

### spec.leaderElectionEnabled

`bool` · optional (explicit presence)

Leader election between operator replicas. Chart default: true.
Meaningful only with replicas above 1 — exactly one replica
reconciles at a time.

- default: `true`

### spec.generateNetworkPolicy

`bool` · optional (explicit presence)

Generate NetworkPolicy resources for the Kafka clusters the
operator manages (broker-to-broker, operator-to-broker and
client-access policies). Chart default: true. Disable on clusters
whose CNI does not enforce NetworkPolicy or where policies are
managed externally — the generated policies are correct but
inert on a non-enforcing CNI.

- default: `true`

### spec.generatePodDisruptionBudget

`bool` · optional (explicit presence)

Generate PodDisruptionBudget resources for the Kafka pods the
operator manages. Chart default: true. On a single-node cluster a
generated PDB can block node drains of a 1-replica pool — disable
only with that trade understood.

- default: `true`

### spec.createGlobalResources

`bool` · optional (explicit presence)

Create the cluster-scoped RBAC the operator needs (ClusterRoles +
bindings). Chart default: true. Set false only for a SECOND
operator install in the same cluster — the fixed-name ClusterRoles
are owned by the first release, and Helm fails a second install
that tries to create them again.

- default: `true`

### spec.resources

`ContainerResources`

Operator container resources. Empty = the chart defaults (requests
200m/384Mi, limits 1000m/384Mi).

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
pulling images from a private mirror.

### spec.image

`KubernetesStrimziKafkaOperatorImage`

Override the registry/organization ALL Strimzi images are pulled
from (the operator itself AND every operand image it deploys:
Kafka, entity operators, Cruise Control, exporter). The air-gap /
private-mirror path. Empty = the chart defaults
(quay.io/strimzi at the chart version).

### spec.image.registry

`string`

Image registry (the chart's defaultImageRegistry, e.g.
"my-mirror.example.com"). Empty = "quay.io".

### spec.image.repository

`string`

Image repository/organization within the registry (the chart's
defaultImageRepository). Empty = "strimzi".

### spec.image.tag

`string`

Image tag (the chart's defaultImageTag). Empty = the chart
version.

### spec.helmValues

`string`

Escape hatch: additional chart values as a YAML document, merged
LAST over everything the typed fields render (Helm `-f` semantics,
identical on both engines). For the chart surface beyond the typed
fields (per-component image overrides, extra env vars, security
contexts, the operator's own PDB/NetworkPolicy, dashboards,
aggregate roles, Connect build timeout, ...) — never the
substitute for them. Do not put secrets here.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesStrimziKafkaOperator, name: <resource-name>, fieldPath: status.outputs.<output>}`.

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
