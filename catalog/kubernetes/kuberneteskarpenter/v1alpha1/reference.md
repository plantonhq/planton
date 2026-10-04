# KubernetesKarpenter

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**KubernetesKarpenterSpec** installs the Karpenter node-provisioning
controller from the official OCI-served Helm charts
(`oci://public.ecr.aws/karpenter/karpenter`, with the CRDs from the
companion `karpenter-crd` chart). Karpenter watches for unschedulable pods
and launches RIGHT-SIZED machines for them in seconds — no pre-created
node groups, no per-ASG scaling policies: it picks instance types from the
live catalog to fit the pending pods, consolidates under-used nodes away,
and handles spot interruptions.

WHAT to provision is declared separately: KubernetesKarpenterNodePool
describes the shape and constraints of the fleet, and
KubernetesKarpenterEc2NodeClass describes the cloud-level machine template
(AMIs, subnets, security groups, IAM). This component installs and
configures the ENGINE; an installation without at least one NodePool
provisions nothing.

ONE INSTALLATION PER CLUSTER: Karpenter owns the cluster-wide
`karpenter.sh` label domain, its CRDs, and node lifecycle — a second
controller fleet would fight over every NodeClaim. The Helm release names
are therefore fixed ("karpenter-crd" and "karpenter").

Karpenter is per-cloud upstream: each cloud ships its own controller,
chart, and NodeClass CRD. This kind installs the AWS provider — the
canonical, mature implementation — and carries the cloud-specific
settings in a typed oneof so future providers land as additive arms.

The typed fields below cover the chart's meaningful configuration
surface; `helm_values` remains as the escape hatch for chart values
beyond them (merged last, Helm `-f` semantics, identical on both
engines) — a safety valve, never the primary interface.

## Example

```yaml
# Full-surface test manifest: exercises every typed arm of the spec so the
# offline plan proofs cover what the live lanes may not. Not a realistic
# production shape — kube-system installs would not normally flip alpha
# gates or host networking.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesKarpenter
metadata:
  name: karpenter
spec:
  namespace:
    value: kube-system
  createNamespace: false
  chartVersion: "1.14.0"
  crds:
    install: true
    keepOnUninstall: true
  cluster:
    name: hack-eks
    endpoint: https://ABC123DEF456.gr7.us-east-1.eks.amazonaws.com
    eksControlPlane: true
  aws:
    irsaRoleArn:
      value: arn:aws:iam::123456789012:role/hack-karpenter-controller
    interruptionQueue: hack-karpenter-interruptions
    isolatedVpc: true
    reservedEnis: 1
    enableZonalShift: true
    vmMemoryOverheadPercent: "0.08"
  controller:
    replicas: 2
    resources:
      requests:
        cpu: "1"
        memory: 1Gi
      limits:
        cpu: "1"
        memory: 1Gi
    logLevel: debug
  batching:
    maxDuration: 30s
    idleDuration: 2s
  scheduling:
    preferencePolicy: Ignore
    minValuesPolicy: BestEffort
  featureGates:
    nodeRepair: true
    nodeOverlay: true
    reservedCapacity: true
    spotToSpotConsolidation: true
    staticCapacity: true
    capacityBuffer: true
  controllerScheduling:
    priorityClassName: system-cluster-critical
    nodeSelector:
      node-role.kubernetes.io/management: "true"
    tolerations:
      - key: node-role.kubernetes.io/control-plane
        operator: Exists
        effect: NoSchedule
    hostNetwork: true
  prometheus:
    serviceMonitor: false
  helmValues: |
    revisionHistoryLimit: 5
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.createNamespace` | `bool` |  |  |  |
| `spec.chartVersion` | `string` |  | `1.14.0` |  |
| `spec.crds` | `KubernetesKarpenterCrds` |  |  |  |
| `spec.crds.install` | `bool` |  | `true` |  |
| `spec.crds.keepOnUninstall` | `bool` |  | `true` |  |
| `spec.cluster` | `KubernetesKarpenterCluster` | yes |  |  |
| `spec.cluster.name` | `string` | yes |  |  |
| `spec.cluster.endpoint` | `string` |  |  |  |
| `spec.cluster.eksControlPlane` | `bool` |  |  |  |
| `spec.cluster.caBundle` | `string` |  |  |  |
| `spec.aws` | `KubernetesKarpenterAws` |  |  |  |
| `spec.aws.irsaRoleArn` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.aws.interruptionQueue` | `string` |  |  |  |
| `spec.aws.isolatedVpc` | `bool` |  |  |  |
| `spec.aws.reservedEnis` | `int32` |  | `0` |  |
| `spec.aws.enableZonalShift` | `bool` |  |  |  |
| `spec.aws.vmMemoryOverheadPercent` | `string` |  | `0.075` |  |
| `spec.controller` | `KubernetesKarpenterController` |  |  |  |
| `spec.controller.replicas` | `int32` |  | `2` |  |
| `spec.controller.resources` | `ContainerResources` |  |  |  |
| `spec.controller.resources.limits` | `CpuMemory` |  |  |  |
| `spec.controller.resources.limits.cpu` | `string` |  |  |  |
| `spec.controller.resources.limits.memory` | `string` |  |  |  |
| `spec.controller.resources.requests` | `CpuMemory` |  |  |  |
| `spec.controller.resources.requests.cpu` | `string` |  |  |  |
| `spec.controller.resources.requests.memory` | `string` |  |  |  |
| `spec.controller.logLevel` | `string` |  | `info` |  |
| `spec.batching` | `KubernetesKarpenterBatching` |  |  |  |
| `spec.batching.maxDuration` | `string` |  | `10s` |  |
| `spec.batching.idleDuration` | `string` |  | `1s` |  |
| `spec.scheduling` | `KubernetesKarpenterSchedulingPosture` |  |  |  |
| `spec.scheduling.preferencePolicy` | `string` |  | `Respect` |  |
| `spec.scheduling.minValuesPolicy` | `string` |  | `Strict` |  |
| `spec.featureGates` | `KubernetesKarpenterFeatureGates` |  |  |  |
| `spec.featureGates.nodeRepair` | `bool` |  |  |  |
| `spec.featureGates.nodeOverlay` | `bool` |  |  |  |
| `spec.featureGates.reservedCapacity` | `bool` |  | `true` |  |
| `spec.featureGates.spotToSpotConsolidation` | `bool` |  |  |  |
| `spec.featureGates.staticCapacity` | `bool` |  |  |  |
| `spec.featureGates.capacityBuffer` | `bool` |  |  |  |
| `spec.controllerScheduling` | `KubernetesKarpenterControllerScheduling` |  |  |  |
| `spec.controllerScheduling.priorityClassName` | `string` |  | `system-cluster-critical` |  |
| `spec.controllerScheduling.nodeSelector` | `map<string, string>` |  |  |  |
| `spec.controllerScheduling.tolerations` | `[]WorkloadToleration` |  |  |  |
| `spec.controllerScheduling.tolerations[].key` | `string` |  |  |  |
| `spec.controllerScheduling.tolerations[].operator` | `string` |  |  |  |
| `spec.controllerScheduling.tolerations[].value` | `string` |  |  |  |
| `spec.controllerScheduling.tolerations[].effect` | `string` |  |  |  |
| `spec.controllerScheduling.tolerations[].tolerationSeconds` | `int64` |  |  |  |
| `spec.controllerScheduling.hostNetwork` | `bool` |  |  |  |
| `spec.prometheus` | `KubernetesKarpenterPrometheus` |  |  |  |
| `spec.prometheus.serviceMonitor` | `bool` |  |  |  |
| `spec.helmValues` | `string` |  |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

Namespace to install Karpenter into ("kube-system" is upstream's
recommended home since v1 — it keeps the controller under the
system-critical eviction umbrella). Accepts a literal namespace name or
a reference to a KubernetesNamespace resource.

Treat the namespace as PERMANENT for a given cluster: the CRDs survive
uninstall by design (so removing Karpenter never deletes the fleet),
and kept CRDs pin the Helm release's namespace in their ownership
metadata — re-installing into a DIFFERENT namespace then fails with
Helm's release-ownership error on CRDs that "should not exist"
(verified live). Moving an existing install requires first deleting
the kept CRDs (only safe with an empty fleet: no NodePools,
EC2NodeClasses, or NodeClaims).

- references: KubernetesNamespace (`spec.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.createNamespace

`bool`

When true, the namespace is created (with the standard Planton
governance labels) before installing and deleted with the resource.
When false, the namespace must already exist. Leave false when
installing into kube-system.

### spec.chartVersion

`string` · optional (explicit presence)

Helm chart version to install (e.g. "1.14.0" — the karpenter and
karpenter-crd charts version together with the controller). Pin
deliberately; upgrades re-run both releases with the new charts. Pick
versions from the OCI registry's published tags
(`oci://public.ecr.aws/karpenter`): the served chart is the contract —
the upstream source tree's Chart.yaml can lag the published tags.

- default: `1.14.0`

### spec.crds

`KubernetesKarpenterCrds`

Karpenter custom resource definitions (NodePool, NodeClaim,
EC2NodeClass, ...) lifecycle — installed through the companion
`karpenter-crd` chart as its own release, upstream's supported
mechanism for keeping CRDs upgradable (Helm never upgrades CRDs
bundled inside the main chart).

- rule: keep_on_uninstall only applies when this resource installs the CRDs — with install false there is nothing to keep

### spec.crds.install

`bool` · optional (explicit presence)

Install the Karpenter CRDs as a dedicated `karpenter-crd` release
before the controller chart. Chart default posture: true. Disable
only when something else manages them — the controller chart bundles
copies Helm installs once and NEVER upgrades, so leaving this on is
what keeps CRDs current across chart upgrades.

- default: `true`

### spec.crds.keepOnUninstall

`bool` · optional (explicit presence)

Keep the CRDs (and therefore every NodePool/EC2NodeClass/NodeClaim in
the cluster) when this resource is destroyed. Rendered as the
`helm.sh/resource-policy: keep` annotation through the CRD chart's
additionalAnnotations — the karpenter-crd chart serves CRDs as
ordinary templates, so WITHOUT this a plain uninstall cascade-deletes
every fleet declaration and Karpenter-provisioned NodeClaim record in
the cluster. Leave true unless you deliberately want uninstall to
purge them.

- default: `true`

### spec.cluster

`KubernetesKarpenterCluster` · required

Identity of the cluster Karpenter provisions for. Required — the
controller cannot start without knowing which cluster's pods and
nodes it owns.

- rule: {"required":true}

### spec.cluster.name

`string` · required

Cluster name — the value nodes register under and (on EKS) the name
used for control-plane discovery. The chart refuses to render without
it.

- rule: {"required":true}

### spec.cluster.endpoint

`string`

Kubernetes API endpoint URL for provisioned nodes to join. Leave
empty on EKS — the controller discovers it at startup (DescribeCluster).
Required in practice for non-EKS control planes.

- rule: endpoint must be an https URL (the cluster API server address, e.g. https://ABC123.gr7.us-east-1.eks.amazonaws.com)

### spec.cluster.eksControlPlane

`bool`

Marks the control plane as EKS, letting Karpenter discover cluster
details (endpoint, CA) from the DescribeCluster API instead of
requiring them here.

### spec.cluster.caBundle

`string`

Cluster CA bundle (base64 PEM) for TLS bootstrap of provisioned
nodes. Empty = taken from the controller's own API-server TLS
configuration — correct for almost every cluster.

### spec.aws

`KubernetesKarpenterAws`

AWS integration: EC2 provisioning, interruption handling, and the
IAM identity the controller runs as.

### spec.aws.irsaRoleArn

`string | valueFrom`

IAM role ARN for IRSA: annotates Karpenter's service account so the
controller calls EC2/EKS/SQS/Pricing without stored keys (the role's
trust policy must allow the cluster's OIDC provider and the
"karpenter" service account). Reference an AwsIamRole's role_arn output
-- the reference is also the deploy-ordering edge, so the role exists
before the controller starts -- or pass a literal ARN
(arn:aws:iam::<account>:role/<name>). Leave unset when using EKS Pod
Identity — the association is configured on the AWS side and needs no
annotation here.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.aws.interruptionQueue

`string`

Name of the SQS queue receiving EC2 interruption events (spot
interruptions, scheduled maintenance, instance rebalance). Empty
disables interruption handling — provisioning still works, but
Karpenter cannot drain ahead of an interruption. The controller role
needs the queue permissions from upstream's IAM guidance.

### spec.aws.isolatedVpc

`bool`

The cluster runs in a VPC without internet-reachable AWS endpoints
beyond the provisioned VPC endpoints. Karpenter then avoids services
without a VPC endpoint — including the pricing API, so price-aware
consolidation falls back to static data.

### spec.aws.reservedEnis

`int32` · optional (explicit presence)

Number of ENIs per node reserved outside Karpenter's max-pods and
kube-reserved math — used with VPC CNI custom networking, where one
ENI per node belongs to the CNI. Chart default: 0.

- default: `0`
- rule: {"int32":{"gte":0}}

### spec.aws.enableZonalShift

`bool`

Respect AWS zonal-shift signals when placing NodeClaims (skips zones
a shift has drained). Requires the cluster to be enabled for zonal
shift and the controller role to carry the ARC permissions.

### spec.aws.vmMemoryOverheadPercent

`string` · optional (explicit presence)

VM memory overhead subtracted from every instance type's reported
memory, as a fraction (0.075 = 7.5% — the chart default). Tune only
when nodes register with less allocatable memory than Karpenter
expected.

- default: `0.075`
- rule: vm_memory_overhead_percent must be a decimal fraction between 0 and 1 (e.g. "0.075")

### spec.controller

`KubernetesKarpenterController`

Controller deployment sizing. Extra replicas are leader-elected warm
standbys, not horizontal capacity.

### spec.controller.replicas

`int32` · optional (explicit presence)

Replica count. Chart default: 2 (an active leader plus a warm
standby, spread across zones by the chart's topology constraints).

- default: `2`
- rule: {"int32":{"gte":1}}

### spec.controller.resources

`ContainerResources`

Container resources. Empty = no requests/limits (upstream leaves the
choice to the operator; 1 CPU / 1Gi is the commonly documented
starting point for large clusters).

### spec.controller.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.controller.resources.limits.cpu

`string`

### spec.controller.resources.limits.memory

`string`

### spec.controller.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.controller.resources.requests.cpu

`string`

### spec.controller.resources.requests.memory

`string`

### spec.controller.logLevel

`string` · optional (explicit presence)

Global log level. Chart default: "info".

- default: `info`
- rule: log_level must be one of 'debug', 'info' or 'error'

### spec.batching

`KubernetesKarpenterBatching`

Pod-batching windows: how long Karpenter waits to gather pending pods
before computing a provisioning decision. Longer windows consider more
pods at once and usually produce fewer, larger nodes.

### spec.batching.maxDuration

`string` · optional (explicit presence)

Maximum batch window (Go duration, e.g. "10s" — the chart default).
The ceiling on how long pending pods wait for a provisioning
decision.

- default: `10s`
- rule: max_duration must be a Go duration such as "10s" or "1m"

### spec.batching.idleDuration

`string` · optional (explicit presence)

Idle window that closes a batch early: if no new pods arrive for this
long, the batch is provisioned immediately (Go duration, chart
default "1s").

- default: `1s`
- rule: idle_duration must be a Go duration such as "1s"

### spec.scheduling

`KubernetesKarpenterSchedulingPosture`

Scheduler-simulation posture: how Karpenter treats soft scheduling
preferences and minValues when computing node shapes.

### spec.scheduling.preferencePolicy

`string` · optional (explicit presence)

How preferred (soft) node/pod affinities and ScheduleAnyways topology
constraints influence provisioning: "Respect" (chart default — soft
preferences shape the node) or "Ignore" (only hard requirements
count, which can pick cheaper shapes).

- default: `Respect`
- rule: preference_policy must be either 'Respect' or 'Ignore'

### spec.scheduling.minValuesPolicy

`string` · optional (explicit presence)

How requirement minValues are treated: "Strict" (chart default —
scheduling fails when minValues cannot be met) or "BestEffort" (relax
minValues rather than leave pods pending).

- default: `Strict`
- rule: min_values_policy must be either 'Strict' or 'BestEffort'

### spec.featureGates

`KubernetesKarpenterFeatureGates`

Alpha/beta feature gates, mirroring the chart's featureGates block.
Gates follow upstream's graduation process — flipping one on opts into
pre-GA behavior.

### spec.featureGates.nodeRepair

`bool`

ALPHA: automatically replace nodes that fail health checks.

### spec.featureGates.nodeOverlay

`bool`

ALPHA: NodeOverlay resources influence scheduling simulations.

### spec.featureGates.reservedCapacity

`bool` · optional (explicit presence)

BETA (chart default true): native on-demand capacity-reservation
support — NodeClasses can select ODCRs and Karpenter launches into
them first.

- default: `true`

### spec.featureGates.spotToSpotConsolidation

`bool`

ALPHA: allow consolidation to replace spot nodes with cheaper spot
nodes (off by default because it can increase churn).

### spec.featureGates.staticCapacity

`bool`

ALPHA: static capacity — NodePools with a fixed replica count instead
of demand-driven provisioning.

### spec.featureGates.capacityBuffer

`bool`

ALPHA: CapacityBuffer resources pre-provision spare headroom.

### spec.controllerScheduling

`KubernetesKarpenterControllerScheduling`

Scheduling of the controller pods themselves (where Karpenter runs —
NOT what it provisions). The chart already pins controller pods away
from Karpenter-provisioned nodes so the controller never disrupts its
own machine.

### spec.controllerScheduling.priorityClassName

`string` · optional (explicit presence)

PriorityClass for the controller pods. Chart default:
"system-cluster-critical" — the provisioner must outlive the
workloads it provisions for.

- default: `system-cluster-critical`

### spec.controllerScheduling.nodeSelector

`map<string, string>`

Node selector for the controller pods. The chart always adds
kubernetes.io/os=linux; entries here narrow it further (e.g. a
dedicated management node group — remember Karpenter must run on
capacity it does NOT manage).

### spec.controllerScheduling.tolerations

`[]WorkloadToleration`

Tolerations for the controller pods. Chart default tolerates
CriticalAddonsOnly.

### spec.controllerScheduling.tolerations[].key

`string`

Taint key to tolerate. Empty key with operator "Exists" tolerates every taint.

### spec.controllerScheduling.tolerations[].operator

`string`

How key/value match: "Equal" (default — value must match too) or "Exists"
(key presence alone matches).

- rule: Toleration operator must be either "Equal" or "Exists"

### spec.controllerScheduling.tolerations[].value

`string`

Taint value to match when operator is "Equal".

### spec.controllerScheduling.tolerations[].effect

`string`

Which taint effect is tolerated: "NoSchedule", "PreferNoSchedule", or
"NoExecute". Empty tolerates all effects for the key.

- rule: Toleration effect must be one of "NoSchedule", "PreferNoSchedule", or "NoExecute"

### spec.controllerScheduling.tolerations[].tolerationSeconds

`int64` · optional (explicit presence)

For "NoExecute" taints only: how many seconds already-running pods stay bound
after the taint appears. Unset means tolerate forever.

### spec.controllerScheduling.hostNetwork

`bool`

Run the controller on the host network — required when the cluster's
CNI cannot serve pod IPs before Karpenter runs (the chicken-and-egg
of Karpenter provisioning the nodes the CNI runs on).

### spec.prometheus

`KubernetesKarpenterPrometheus`

Karpenter's own Prometheus telemetry.

### spec.prometheus.serviceMonitor

`bool`

Create a ServiceMonitor for scrape discovery of the controller's
metrics endpoint (provisioning latencies, disruption decisions,
pricing lookups). Requires the Prometheus operator CRDs on the
cluster — the release FAILS to install without them.

### spec.helmValues

`string`

Escape hatch: additional chart values as a YAML document, merged LAST
over everything the typed fields render (Helm `-f` semantics,
identical on both engines). For the chart surface beyond the typed
fields (image digests/registries, extra volumes, sidecars, DNS
config, pod disruption budget tuning, ...) — never the substitute for
them. Do not put secrets here.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesKarpenter, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | Namespace Karpenter was installed into (the resolved spec.namespace). |
| `status.outputs.release_name` | `string` | Controller Helm release name — fixed "karpenter" (one installation per cluster; the controller owns node lifecycle and the karpenter.sh label domain). |
| `status.outputs.crd_release_name` | `string` | CRD Helm release name — fixed "karpenter-crd" (empty when spec.crds.install is false and something else manages the CRDs). |
| `status.outputs.service_account_name` | `string` | Name of the controller's Kubernetes service account (the chart's fixed "karpenter") — the subject IRSA trust policies and EKS Pod Identity associations are written against. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.aws.irsaRoleArn` | AwsIamRole | `status.outputs.role_arn` |

## See Also

- [Overview](../README.md)
