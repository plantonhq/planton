# KubernetesFlinkOperator

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**KubernetesFlinkOperatorSpec** deploys the Apache Flink Kubernetes
Operator — the official ASF controller that turns
KubernetesFlinkDeployment declarations (and, unmodeled here,
FlinkSessionJob/FlinkStateSnapshot/FlinkBlueGreenDeployment CRs
authored directly) into running Flink clusters — from the official
chart served per version at
https://downloads.apache.org/flink/flink-kubernetes-operator-<version>/.

GRAIN: one operator per namespace by construction — the chart
hardcodes its webhook Service, certificate, and issuer names
(`flink-operator-webhook-service`, `flink-operator-serving-cert`),
so a second release in the same namespace collides. One
cluster-wide-watching operator is the normal posture.

THE WEBHOOK LIFECYCLE, READ THIS FIRST (chart truth at 1.15.0):
with the webhook enabled (the upstream default this spec keeps),
the chart renders cert-manager Issuer/Certificate resources
UNCONDITIONALLY and the webhook trusts the API server through
cert-manager's CA injection — there is NO self-signed fallback,
which makes KubernetesCertManager a hard prerequisite. Both
webhook configurations are FAIL-CLOSED: if the webhook cannot be
reached (cert-manager absent, operator down), EVERY
flink.apache.org admission in scope is rejected — a policy-engine
blast radius, not a soft degradation. `webhook_enabled: false`
removes the webhook, the certificate machinery, and the
cert-manager dependency; the operator still validates in its
reconcile loop (failures surface on CR status instead of at
admission).

The chart's default webhook keystore credential is a HARDCODED
PUBLIC PASSWORD — it never ships from this component: the modules
generate a random keystore password Secret per install.

CRD LIFECYCLE: the chart ships its four flink.apache.org CRDs in
its `crds/` directory — Helm installs them once, NEVER upgrades
them on chart upgrades, and LEAVES them (and every Flink
declaration) on uninstall. Apply the new release's CRD files
manually when a chart bump changes them.

OPERATOR CONFIGURATION rides Flink's own config format
(`kubernetes.operator.*` keys appended over the chart defaults) —
see `operator_config`. Job/cluster-level Flink config belongs on
each KubernetesFlinkDeployment, not here — entries here become
cluster-wide defaults for every deployment the operator manages.

NAMING: keep the resource name at 45 characters or fewer — the
modules pin the chart's fullname to the resource name, the longest
derived name is the module-generated "-webhook-keystore" Secret
(17 chars — the chart's own webhook artifact names are FIXED, not
name-derived), and Kubernetes names cap at 63. Both engines fail
loudly over the budget.

The typed fields cover the chart's meaningful surface;
`helm_values` remains the escape hatch (merged last, Helm `-f`
semantics, identical on both engines) — logging framework,
operator volume mounts, the NodePort RBAC rule — a safety valve,
never the primary interface.

## Example

```yaml
# Full-surface hack manifest for the offline plan/preview proofs: every
# module-rendered arm of the operator expressed at once — the fenced
# watch posture, the default-on webhook (exercising the module-generated
# keystore password Secret and the useDefaultPassword=false re-pin),
# leader-elected replicas (exercising the module-owned leader-election
# config), operator config appends, a custom job service account,
# resources, the image-registry mirror (with the always-rendered tag
# pin), pull secrets, full scheduling, and an escape-hatch document
# (logging framework).
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesFlinkOperator
metadata:
  name: flink-operator-full
spec:
  namespace:
    value: flink-system
  createNamespace: true
  chartVersion: 1.15.0
  watchNamespaces:
    - stream-team-a
    - stream-team-b
  webhookEnabled: true
  replicas: 2
  operatorConfig:
    kubernetes.operator.reconcile.interval: 15 s
    kubernetes.operator.metrics.reporter.slf4j.interval: 5 MINUTE
  jobServiceAccount: flink
  resources:
    requests:
      cpu: 250m
      memory: 512Mi
    limits:
      cpu: "1"
      memory: 1Gi
  imageRegistry: mirror.example.com
  imagePullSecrets:
    - mirror-pull
  scheduling:
    nodeSelector:
      workload: system
    tolerations:
      - key: dedicated
        operator: Equal
        value: stream-platform
        effect: NoSchedule
    priorityClassName: system-cluster-critical
  helmValues: |
    logging:
      framework: log4j2
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.createNamespace` | `bool` |  |  |  |
| `spec.chartVersion` | `string` |  | `1.15.0` |  |
| `spec.watchNamespaces` | `[]string` |  |  |  |
| `spec.webhookEnabled` | `bool` |  | `true` |  |
| `spec.replicas` | `int32` |  | `1` |  |
| `spec.operatorConfig` | `map<string, string>` |  |  |  |
| `spec.jobServiceAccount` | `string` |  | `flink` |  |
| `spec.resources` | `ContainerResources` |  |  |  |
| `spec.resources.limits` | `CpuMemory` |  |  |  |
| `spec.resources.limits.cpu` | `string` |  |  |  |
| `spec.resources.limits.memory` | `string` |  |  |  |
| `spec.resources.requests` | `CpuMemory` |  |  |  |
| `spec.resources.requests.cpu` | `string` |  |  |  |
| `spec.resources.requests.memory` | `string` |  |  |  |
| `spec.imageRegistry` | `string` |  |  |  |
| `spec.imagePullSecrets` | `[]string` |  |  |  |
| `spec.scheduling` | `KubernetesFlinkOperatorScheduling` |  |  |  |
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

Operator version to install (e.g. "1.15.0"). The chart is served
per version from
https://downloads.apache.org/flink/flink-kubernetes-operator-<version>/
(chart version = operator version = image tag; the modules pin
the image tag explicitly — the chart's own default is the
unpinned "latest"). NOTE chart upgrades never upgrade the
flink.apache.org CRDs (`crds/` directory posture) — apply new
CRD files manually when a bump changes them.

- default: `1.15.0`

### spec.watchNamespaces

`[]string`

Namespaces the operator watches for Flink CRs. Empty = every
namespace (the normal one-operator-per-cluster posture). With a
list set, the chart scopes RBAC AND the admission webhook to
exactly these namespaces — Flink declarations outside them are
ignored without an error. The modules CREATE each listed
namespace before the Helm release: the chart plants job
ServiceAccount/Role/RoleBinding INTO those namespaces and does
not create them — an absent watch namespace fails the install
with "namespaces \"…\" not found".

### spec.webhookEnabled

`bool` · optional (explicit presence)

The admission webhook (validates AND defaults Flink CRs at
admission time). Empty = true — the upstream default, and the
reason KubernetesCertManager is this component's prerequisite
(see the spec-level WEBHOOK LIFECYCLE note: fail-closed, CA
injection, no self-signed arm). Set false to run without
webhook, certificates, or the cert-manager dependency —
validation then happens in the reconcile loop.

- default: `true`

### spec.replicas

`int32` · optional (explicit presence)

Number of operator replicas. More than 1 requires leader
election — the modules render the operator's leader-election
config for you (one active reconciler, warm standbys); without
it the chart REFUSES multi-replica installs by design.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.operatorConfig

`map<string, string>`

Operator configuration appended over the chart defaults —
Flink config keys (e.g.
"kubernetes.operator.reconcile.interval": "15 s",
autoscaler defaults, metrics reporters). Entries become
cluster-wide defaults for every FlinkDeployment this operator
manages; per-deployment config belongs on the deployment.

### spec.jobServiceAccount

`string` · optional (explicit presence)

Service account name Flink JOB pods run as (the chart creates it
with reconcile RBAC in the operator's namespace and, with
watch_namespaces set, in each watched namespace). Empty =
"flink" — the name every FlinkDeployment references by default.
The chart marks it `helm.sh/resource-policy: keep`: it survives
uninstall so running jobs never lose their identity.

- default: `flink`

### spec.resources

`ContainerResources`

CPU and memory for the operator container. Empty = the chart
defaults. The operator is a JVM — production installs typically
set requests explicitly.

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
(`ghcr.io/apache/flink-kubernetes-operator`) — the
air-gap/private-mirror path. This does NOT rewrite the Flink
images deployments run — those ride each
KubernetesFlinkDeployment's own image field.

### spec.imagePullSecrets

`[]string`

Names of existing image-pull Secrets applied to the operator
pods. The Secrets must already exist in the namespace.

### spec.scheduling

`KubernetesFlinkOperatorScheduling`

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
beyond the typed fields (logging framework, operator volumes,
health-probe tuning, JVM args) — never the substitute for them.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesFlinkOperator, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | namespace the operator runs in. |
| `status.outputs.release_name` | `string` | Helm release name (= metadata.name). The modules pin the chart's fullname to it. |
| `status.outputs.job_service_account` | `string` | service account name Flink job pods run as — FlinkDeployment declarations reference it. |
| `status.outputs.watched_namespaces` | `[]string` | namespaces the operator watches for Flink CRs. Empty means cluster-wide — FlinkDeployments reconcile anywhere. |
| `status.outputs.webhook_service` | `string` | name of the operator's webhook Service (chart-fixed `flink-operator-webhook-service`) — empty when the webhook is disabled. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |

## See Also

- [Overview](../README.md)
